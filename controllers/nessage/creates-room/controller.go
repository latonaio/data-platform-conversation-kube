package controllersMessageCreatesRoom

import (
	"data-platform-conversation-kube/api-input-reader/types"
	"data-platform-conversation-kube/services"
	"github.com/astaxie/beego"
	"github.com/latonaio/golang-logging-library-for-data-platform/logger"
	database "github.com/latonaio/golang-mysql-network-connector"
)

type MessageCreatesRoomController struct {
	beego.Controller
	UserInfo     *types.Request
	CustomLogger *logger.Logger
	DB           *database.Mysql
}

func (controller *MessageCreatesRoomController) Get() {
	roomPartner, _ := controller.GetInt("roomPartner")

	controller.UserInfo = services.UserRequestParams(
		services.RequestWrapperController{
			Controller:   &controller.Controller,
			CustomLogger: controller.CustomLogger,
		},
	)

	chatRoom, err := services.CreateChatRoom(
		controller.DB,
		*controller.UserInfo.BusinessPartner,
		roomPartner,
	)

	if err != nil {
		services.HandleError(
			&controller.Controller,
			err,
			nil,
		)
		controller.CustomLogger.Error("CreateChatRoom error")
		return
	}

	businessPartners := []int{
		*controller.UserInfo.BusinessPartner,
		roomPartner,
	}

	businessPartnerDocImages, err := services.ReadBusinessPartnerDocs(
		controller.DB,
		businessPartners,
	)

	controller.Data["json"] = map[string]interface{}{
		"ChatRoom":                 chatRoom,
		"BusinessPartnerDocImages": businessPartnerDocImages,
	}
	controller.ServeJSON()
}
