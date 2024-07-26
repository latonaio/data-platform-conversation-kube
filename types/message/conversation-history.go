package typesMessage

type ConversationHistoryWithReadStatus struct {
	MessageID       string
	ChatRoom        string
	BusinessPartner int
	Content         string
	SentAt          string
	ReadStatusID    *string
	ReadAt          *string
}
