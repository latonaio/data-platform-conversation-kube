package main

import (
	"data-platform-conversation-kube/config"
	_ "data-platform-conversation-kube/routers"
	"github.com/astaxie/beego"
)

func main() {
	conf := config.NewConf()
	beego.Run(conf.SERVER.ServerURL())
	//beego.Run()
}
