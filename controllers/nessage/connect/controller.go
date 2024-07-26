package controllersMessageConnect

import (
	"data-platform-conversation-kube/services"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/latonaio/golang-logging-library-for-data-platform/logger"
	database "github.com/latonaio/golang-mysql-network-connector"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type MessageConnectController struct {
	beego.Controller
	CustomLogger *logger.Logger
	DB           *database.Mysql
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	rooms = make(map[string]map[string]*websocket.Conn)
	mu    sync.Mutex
)

type Message struct {
	Type          string  `json:"type"`
	Content       *any    `json:"content,omitempty"`
	MessageID     *string `json:"messageID,omitempty"`
	MessageSender *int    `json:"messageSender,omitempty"`
	MessageReader *int    `json:"messageReader,omitempty"`
}

const (
	Error                   = "Error"
	LeftChat                = "LeftChat"
	ReceivedMessage         = "ReceivedMessage"
	MarkedMessageToSender   = "MarkedMessageToSender"
	MarkedMessageFromReader = "MarkedMessageFromReader"
)

const (
	JoinToRoom                                        = "JoinToRoom"
	InsertMessageHistory                              = "InsertMessageHistory"
	SendMessageToReceiver                             = "SendMessageToReceiver"
	SendErrorResponse                                 = "SendErrorResponse"
	ConvertBusinessPartnerIDToInt                     = "ConvertBusinessPartnerIDToInt"
	ConvertMessageReaderToInt                         = "ConvertMessageReaderToInt"
	ConvertMessageReaderToIntToMessageReader          = "ConvertMessageReaderToIntToMessageReader"
	InsertMessageIntoMessageReadStatus                = "InsertMessageIntoMessageReadStatus"
	InsertMessageIntoMessageReadStatusToMessageReader = "InsertMessageIntoMessageReadStatusToMessageReader"
)

var ErrorMessages = map[string]string{
	JoinToRoom:                                        "Failed to join to room",
	InsertMessageHistory:                              "Failed to insert message into history",
	SendMessageToReceiver:                             "Failed to send message to receiver",
	SendErrorResponse:                                 "Failed to send error response",
	ConvertBusinessPartnerIDToInt:                     "Failed to convert businessPartnerID to int",
	ConvertMessageReaderToInt:                         "Failed to convert messageReader to int",
	ConvertMessageReaderToIntToMessageReader:          "Failed to convert messageReader to int to message reader",
	InsertMessageIntoMessageReadStatus:                "Failed to insert message into message read status",
	InsertMessageIntoMessageReadStatusToMessageReader: "Failed to insert message into message read status to message reader",
}

func (controller *MessageConnectController) Connect() {
	chatRoom := controller.GetString(":chatRoom")
	businessPartnerStr := controller.GetString(":businessPartner")
	businessPartner, err := strconv.Atoi(businessPartnerStr)
	if err != nil {
		controller.CustomLogger.Error(
			"Failed to convert businessPartner to int: ",
			err,
			businessPartnerStr,
		)
		return
	}

	ws, err := upgrader.Upgrade(
		controller.Ctx.ResponseWriter,
		controller.Ctx.Request,
		nil,
	)
	if err != nil {
		controller.CustomLogger.Error("Failed to set websocket upgrade:", err, chatRoom, businessPartner)
		return
	}
	defer ws.Close()

	controller.CustomLogger.Info("Connected room id: %s %s", chatRoom, businessPartner)

	mu.Lock()
	if rooms[chatRoom] == nil {
		rooms[chatRoom] = make(map[string]*websocket.Conn)
	}
	rooms[chatRoom][strconv.Itoa(businessPartner)] = ws
	mu.Unlock()

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			controller.CustomLogger.Error(
				"Read error: ",
				err,
				chatRoom,
				businessPartner,
			)
			break
		}
		switch msg.Type {
		case "SendMessage":
			var messageID string
			if msg.MessageID != nil {
				messageID = *msg.MessageID
			} else {
				controller.CustomLogger.Error(
					"MessageID is nil",
					chatRoom,
					businessPartner,
				)
				continue
			}
			var messageContent string
			if msg.Content != nil {
				messageContent = fmt.Sprintf("%v", *msg.Content)
			}

			controller.sendMessage(
				ws,
				rooms[chatRoom],
				chatRoom,
				businessPartner,
				messageID,
				messageContent,
			)
		case "LeaveRoom":
			controller.leaveRoom(ws, chatRoom)
			controller.CustomLogger.Info("Leave room: ", chatRoom, businessPartner)
		case "MarkMessageAsRead":
			var messageSender int
			if msg.MessageSender != nil {
				messageSender = *msg.MessageSender
			} else {
				controller.CustomLogger.Error(
					"MessageSender is nil",
					chatRoom,
					businessPartner,
				)
				continue
			}
			var messageReader int
			if msg.MessageReader != nil {
				messageReader = *msg.MessageReader
			} else {
				controller.CustomLogger.Error(
					"MessageReader is nil",
					chatRoom,
					businessPartner,
				)
				continue
			}
			var messageID string
			if msg.MessageID != nil {
				messageID = *msg.MessageID
			} else {
				controller.CustomLogger.Error(
					"MessageID is nil",
					chatRoom,
					businessPartner,
				)
				continue
			}

			controller.markMessageAsRead(
				ws,
				rooms[chatRoom],
				chatRoom,
				messageSender,
				messageReader,
				messageID,
			)
		}
	}

	controller.disconnect(rooms[chatRoom], chatRoom, businessPartner)
}

func (controller *MessageConnectController) sendMessage(
	ws *websocket.Conn,
	roomConnections map[string]*websocket.Conn,
	chatRoom string,
	businessPartner int,
	messageID string,
	content string,
) {
	sentAt := time.Now().Format("2006-01-02 15:04:05.999999")

	err := services.InsertConversationHistory(
		controller.DB,
		chatRoom, businessPartner,
		messageID, content,
		sentAt,
	)
	if err != nil {
		controller.CustomLogger.Error(
			ErrorMessages[InsertMessageHistory],
			err,
			messageID, chatRoom, businessPartner,
		)
		err = ws.WriteJSON(map[string]any{
			"type":      Error,
			"message":   ErrorMessages[InsertMessageHistory],
			"messageID": messageID,
			"chatRoom":  chatRoom,
			"sender":    businessPartner,
			"sentAt":    sentAt,
		})
		if err != nil {
			controller.CustomLogger.Error(
				"Failed to send error response: ",
				err,
				messageID, chatRoom, businessPartner,
			)
		}
		return
	}

	for _, ws := range roomConnections {
		go func(ws *websocket.Conn) {
			err = ws.WriteJSON(map[string]any{
				"type":      ReceivedMessage,
				"messageID": messageID,
				"content":   content,
				"chatRoom":  chatRoom,
				"sender":    businessPartner,
				"sentAt":    sentAt,
			})
			if err != nil {
				controller.CustomLogger.Error(
					ErrorMessages[SendMessageToReceiver],
					err,
					messageID, chatRoom, businessPartner,
				)
				err = ws.WriteJSON(map[string]any{
					"type":      Error,
					"message":   ErrorMessages[SendMessageToReceiver],
					"messageID": messageID,
					"chatRoom":  chatRoom,
					"sender":    businessPartner,
					"sentAt":    sentAt,
				})
				if err != nil {
					controller.CustomLogger.Error(
						ErrorMessages[SendErrorResponse],
						err,
						messageID, chatRoom, businessPartner,
					)
				}
			}
		}(ws)
	}
}

func (controller *MessageConnectController) leaveRoom(ws *websocket.Conn, chatRoom string) {
	mu.Lock()
	defer mu.Unlock()
	delete(rooms, chatRoom)
	fmt.Printf("left room %s\n", chatRoom)
	for room := range rooms {
		if room == chatRoom {
			err := ws.WriteJSON(map[string]any{
				"type":     LeftChat,
				"message":  fmt.Sprintf("RoomID %s left the chat", chatRoom),
				"chatRoom": chatRoom,
				//"businessPartner": businessPartner,
			})
			if err != nil {
				controller.CustomLogger.Error(
					"Failed to send left chat error message: ",
					err,
					chatRoom,
					//businessPartner,
				)
				ws.WriteJSON(map[string]any{
					"type":    Error,
					"message": "Failed to send left chat error message",
				})
			}
		}
	}
}

func (controller *MessageConnectController) markMessageAsRead(
	ws *websocket.Conn,
	roomConnections map[string]*websocket.Conn,
	roomID string,
	messageSender int,
	messageReader int,
	messageID string,
) {
	readAt := time.Now().Format("2006-01-02 15:04:05.999999")
	readStatusID := uuid.New().String()

	err := services.InsertMessageReadStatus(
		controller.DB,
		readStatusID,
		messageID,
		messageReader,
		readAt,
	)
	if err != nil {
		controller.CustomLogger.Error(
			ErrorMessages[InsertMessageIntoMessageReadStatus],
			err,
			messageID, roomID, messageSender, messageReader,
			readStatusID, readAt,
		)
		err = ws.WriteJSON(map[string]any{
			"type":          Error,
			"message":       ErrorMessages[InsertMessageIntoMessageReadStatus],
			"messageID":     messageID,
			"roomID":        roomID,
			"messageSender": messageSender,
			"messageReader": messageReader,
			"readStatusID":  readStatusID,
			"readAt":        readAt,
		})
		if err != nil {
			controller.CustomLogger.Error(
				ErrorMessages[SendErrorResponse],
				err,
				messageID, roomID, messageSender, messageReader,
				readStatusID, readAt,
			)
		}
		return
	}

	for roomConnectorBusinessPartnerID, ws := range roomConnections {
		parsedBusinessPartnerID, err := strconv.Atoi(roomConnectorBusinessPartnerID)

		if err != nil {
			controller.CustomLogger.Error(
				ErrorMessages[ConvertBusinessPartnerIDToInt],
				err,
				messageID, roomID, messageSender, messageReader,
				readStatusID, readAt,
			)
			return
		}

		if parsedBusinessPartnerID == messageSender {
			go func(ws *websocket.Conn) {
				err = ws.WriteJSON(map[string]any{
					"type":         MarkedMessageToSender,
					"roomID":       roomID,
					"messageID":    messageID,
					"readStatusID": readStatusID,
					"readAt":       readAt,
				})
			}(ws)
		} else if parsedBusinessPartnerID == messageReader {
			go func(ws *websocket.Conn) {
				err = ws.WriteJSON(map[string]any{
					"type":         MarkedMessageFromReader,
					"roomID":       roomID,
					"messageID":    messageID,
					"readStatusID": readStatusID,
					"readAt":       readAt,
				})
			}(ws)
		}
	}
}

func (controller *MessageConnectController) disconnect(
	roomConnections map[string]*websocket.Conn,
	roomID string,
	businessPartner int,
) {
	mu.Lock()
	defer mu.Unlock()
	delete(rooms[roomID], strconv.Itoa(businessPartner))
	if len(rooms[roomID]) == 0 {
		delete(rooms, roomID)
	}

	for _, ws := range roomConnections {
		err := ws.WriteJSON(map[string]any{
			"type":            LeftChat,
			"message":         fmt.Sprintf("Disconnected user %s", businessPartner),
			"roomID":          roomID,
			"businessPartner": businessPartner,
		})
		if err != nil {
			controller.CustomLogger.Error(
				"Failed to send to disconnected message: ",
				err,
				roomID,
				businessPartner,
			)
			ws.WriteJSON(map[string]any{
				"type":    Error,
				"message": "Failed to send to disconnected message",
			})
		}
	}
	controller.CustomLogger.Info("Disconnected: %s %s", roomID, businessPartner)
}
