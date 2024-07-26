package controllersMessageUserProfile

import (
	"data-platform-conversation-kube/api-input-reader/types"
	"data-platform-conversation-kube/services"
	"github.com/astaxie/beego"
	"github.com/latonaio/golang-logging-library-for-data-platform/logger"
	database "github.com/latonaio/golang-mysql-network-connector"
)

type MessageUserProfileController struct {
	beego.Controller
	UserInfo     *types.Request
	CustomLogger *logger.Logger
	DB           *database.Mysql
}

func (controller *MessageUserProfileController) Get() {
	businessPartner, _ := controller.GetInt(":businessPartner")

	controller.UserInfo = services.UserRequestParams(
		services.RequestWrapperController{
			Controller:   &controller.Controller,
			CustomLogger: controller.CustomLogger,
		},
	)

	userProfile, err := services.ReadBusinessPartnerWithDetails(
		controller.DB,
		businessPartner,
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

	controller.Data["json"] = map[string]interface{}{
		"UserProfile": userProfile,
	}
	controller.ServeJSON()
}
