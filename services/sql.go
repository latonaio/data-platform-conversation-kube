package services

import (
	typesMessage "data-platform-conversation-kube/types/message"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	database "github.com/latonaio/golang-mysql-network-connector"
	"strings"
	"time"
)

type BusinessPartnerDoc struct {
	BusinessPartner          int
	DocType                  string
	DocVersionID             int
	DocID                    string
	FileExtension            string
	FileName                 *string
	FilePath                 *string
	DocIssuerBusinessPartner *int
}

func CreateChatRoom(
	db *database.Mysql,
	roomCreator int,
	roomPartner int,
) (*string, error) {
	now := time.Now()
	chatRoom := uuid.New().String()

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()

	checkQuery := `
        SELECT ChatRoom
        FROM data_platform_chat_room_header_data
        WHERE (RoomCreator = ? AND RoomPartner = ?)
           OR (RoomCreator = ? AND RoomPartner = ?)
    `
	var existingRoomID string
	err = tx.QueryRow(
		checkQuery,
		roomCreator,
		roomPartner,
		roomPartner,
		roomCreator,
	).Scan(&existingRoomID)
	if err == nil {
		return &existingRoomID, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	insertQuery := `
        INSERT INTO data_platform_chat_room_header_data (
            ChatRoom,
            RoomCreator,
            RoomPartner,
            CreatedAt,
            UpdatedAt
        ) VALUES (?, ?, ?, ?, ?)
    `

	_, err = tx.Exec(
		insertQuery,
		chatRoom,
		roomCreator,
		roomPartner,
		now,
		now,
	)

	if err != nil {
		return nil, err
	}

	return &chatRoom, nil
}

func ReadConversationHistoryWithReadStatus(
	db *database.Mysql,
	chatRoom string,
) (*[]typesMessage.ConversationHistoryWithReadStatus, error) {
	query := `
        SELECT 
            message.MessageID, 
            message.ChatRoom, 
            message.BusinessPartner, 
            message.Content, 
            CONCAT(DATE_FORMAT(message.SentAt, '%Y-%m-%d %H:%i:%s'),'.',LPAD(FLOOR(MICROSECOND(message.SentAt) / 1000), 3, '0')) AS SentAt,
            messageReadStatus.ReadStatusID,
            CONCAT(DATE_FORMAT(messageReadStatus.ReadAt, '%Y-%m-%d %H:%i:%s'),'.',LPAD(FLOOR(MICROSECOND(messageReadStatus.ReadAt) / 1000), 3, '0')) AS ReadAt
        FROM 
            data_platform_chat_room_message_data AS message
        LEFT JOIN 
            data_platform_chat_room_message_read_status_data AS messageReadStatus
        ON 
            message.MessageID = messageReadStatus.MessageID
        WHERE 
            message.ChatRoom = ?
    `
	rows, err := db.Query(query, chatRoom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var histories []typesMessage.ConversationHistoryWithReadStatus
	for rows.Next() {
		var history typesMessage.ConversationHistoryWithReadStatus
		var readStatusID sql.NullString
		var readAt sql.NullString

		if err := rows.Scan(
			&history.MessageID,
			&history.ChatRoom,
			&history.BusinessPartner,
			&history.Content,
			&history.SentAt,
			&readStatusID,
			&readAt,
		); err != nil {
			return nil, err
		}

		if readStatusID.Valid {
			history.ReadStatusID = &readStatusID.String
		}
		if readAt.Valid {
			history.ReadAt = &readAt.String
		}

		histories = append(histories, history)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &histories, nil
}

func InsertConversationHistory(
	db *database.Mysql,
	chatRoom string,
	businessPartner int,
	messageID string,
	message string,
	sentAt string,
) error {
	insertQuery := `
        INSERT INTO data_platform_chat_room_message_data (
            MessageID,
            ChatRoom,
            BusinessPartner,
            Content,
            SentAt
        ) VALUES (?, ?, ?, ?, ?)
    `
	_, err := db.Exec(insertQuery, messageID, chatRoom, businessPartner, message, sentAt)
	if err != nil {
		return err
	}

	return err
}

func ReadBusinessPartnerDocs(
	db *database.Mysql,
	businessPartners []int,
) (*[]BusinessPartnerDoc, error) {
	placeholders := strings.Repeat("?,", len(businessPartners)-1) + "?"

	query := `
        SELECT BusinessPartner, DocType, DocVersionID, DocID, FileExtension, FileName, FilePath, DocIssuerBusinessPartner
        FROM data_platform_business_partner_general_doc_data
        WHERE BusinessPartner IN (` + placeholders + `)
    `

	rows, err := db.Query(query, toInterfaceSlice(businessPartners)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []BusinessPartnerDoc
	for rows.Next() {
		var doc BusinessPartnerDoc
		err := rows.Scan(
			&doc.BusinessPartner,
			&doc.DocType,
			&doc.DocVersionID,
			&doc.DocID,
			&doc.FileExtension,
			&doc.FileName,
			&doc.FilePath,
			&doc.DocIssuerBusinessPartner,
		)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &docs, nil
}

func InsertMessageReadStatus(
	db *database.Mysql,
	readStatusID string,
	messageID string,
	participant int,
	readAt string,
) error {
	insertQuery := `
        INSERT INTO data_platform_chat_room_message_read_status_data (
            ReadStatusID,
            MessageID,
            Participant,
            ReadAt
        ) VALUES (?, ?, ?, ?)
    `
	_, err := db.Exec(insertQuery, readStatusID, messageID, participant, readAt)
	if err != nil {
		return err
	}

	return nil
}

func ReadBusinessPartnerWithDetails(
	db *database.Mysql,
	businessPartnerID int,
) (*[]typesMessage.BusinessPartnerWithDetails, error) {
	query := `
        SELECT
            bp.BusinessPartner,
            bp.BusinessPartnerType,
            bp.NickName,
            bp.ProfileComment,
            bp.PreferableLocalSubRegion,
            bp.PreferableLocalRegion,
            bp.PreferableCountry,
            lrtd.LocalRegionName,
            lsrt.LocalSubRegionName
        FROM
            data_platform_business_partner_person_data AS bp
        LEFT JOIN
            data_platform_local_region_text_data AS lrtd
        ON
            bp.PreferableLocalRegion = lrtd.LocalRegion
            AND bp.PreferableCountry = lrtd.Country
            AND bp.Language = lrtd.Language
        LEFT JOIN
            data_platform_local_sub_region_text_data AS lsrt
        ON
            bp.PreferableLocalSubRegion = lsrt.LocalSubRegion
            AND bp.PreferableLocalRegion = lsrt.LocalRegion
            AND bp.PreferableCountry = lsrt.Country
            AND bp.Language = lsrt.Language
        WHERE
            bp.BusinessPartner = ?
    `

	rows, err := db.Query(query, businessPartnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var partners []typesMessage.BusinessPartnerWithDetails
	for rows.Next() {
		var partner typesMessage.BusinessPartnerWithDetails
		var profileComment sql.NullString
		var localRegionName sql.NullString
		var localSubRegionName sql.NullString

		if err := rows.Scan(
			&partner.BusinessPartner,
			&partner.BusinessPartnerType,
			&partner.NickName,
			&profileComment,
			&partner.PreferableLocalSubRegion,
			&partner.PreferableLocalRegion,
			&partner.PreferableCountry,
			&localRegionName,
			&localSubRegionName,
		); err != nil {
			return nil, err
		}

		if profileComment.Valid {
			partner.ProfileComment = &profileComment.String
		}
		if localRegionName.Valid {
			partner.LocalRegionName = &localRegionName.String
		}
		if localSubRegionName.Valid {
			partner.LocalSubRegionName = &localSubRegionName.String
		}

		partners = append(partners, partner)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &partners, nil
}

func toInterfaceSlice(ints []int) []interface{} {
	interfaces := make([]interface{}, len(ints))
	for i, v := range ints {
		interfaces[i] = v
	}
	return interfaces
}
