package routers

import (
	"data-platform-conversation-kube/config"
	"data-platform-conversation-kube/controllers/nessage/connect"
	"data-platform-conversation-kube/controllers/nessage/creates-room"
	controllersMessageHistories "data-platform-conversation-kube/controllers/nessage/histories"
	controllersMessageUserProfile "data-platform-conversation-kube/controllers/nessage/user-profile"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/latonaio/golang-logging-library-for-data-platform/logger"
	database "github.com/latonaio/golang-mysql-network-connector"
)

func init() {
	l := logger.NewLogger()
	conf := config.NewConf()
	db, err := database.NewMySQL(conf.DB)
	if err != nil {
		l.Fatal(err.Error())
	}
	l.Info("DB connection established")

	messageConnectController := &controllersMessageConnect.MessageConnectController{
		CustomLogger: l,
		DB:           db,
	}

	messageHistoriesController := &controllersMessageHistories.MessageHistoriesController{
		CustomLogger: l,
		DB:           db,
	}

	messageCreatesRoomController := &controllersMessageCreatesRoom.MessageCreatesRoomController{
		CustomLogger: l,
		DB:           db,
	}

	messageUserProfileController := &controllersMessageUserProfile.MessageUserProfileController{
		CustomLogger: l,
		DB:           db,
	}

	chat := beego.NewNamespace(
		"/message",
		beego.NSCond(func(ctx *context.Context) bool { return true }),
		beego.NSRouter("/creates/room", messageCreatesRoomController),
		beego.NSRouter("/histories/:chatRoom", messageHistoriesController),
		beego.NSRouter("/user-profile/:businessPartner", messageUserProfileController),
		beego.NSRouter("/connect/:chatRoom/:businessPartner", messageConnectController, "get:Connect"),
	)

	beego.AddNamespace(
		beego.NewNamespace("/api/conversation").
			Namespace(
				chat,
			),
	)

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))
}
