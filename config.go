package fetcp

import (
	"github.com/astaxie/beego"
)

type SrvConfig struct {
	ServerName             string
	Port                   int
	PacketSendChanLimit    int
	PacketReceiveChanLimit int
	ConnectTimeOut         int64
	HeatbeatCheck          bool
	HeatbeatCheckSpec      int
}

var (
	ServerConfig *SrvConfig
)

func init() {
	ServerConfig = &SrvConfig{
		ServerName:             "MyServer",
		Port:                   9000,
		PacketSendChanLimit:    4096,
		PacketReceiveChanLimit: 4096,
		ConnectTimeOut:         300,
		HeatbeatCheck:          false,
		HeatbeatCheckSpec:      5,
	}
	ReloadConfig()
}

func ReloadConfig() {
	beego.LoadAppConfig("ini", "./system.ini")
	ServerConfig.ServerName = beego.AppConfig.DefaultString("server::ServerName", ServerConfig.ServerName)
	ServerConfig.Port = beego.AppConfig.DefaultInt("server::Port", ServerConfig.Port)
	ServerConfig.PacketSendChanLimit = beego.AppConfig.DefaultInt("server::PacketSendChanLimit", ServerConfig.PacketSendChanLimit)
	ServerConfig.PacketReceiveChanLimit = beego.AppConfig.DefaultInt("server::PacketReceiveChanLimit", ServerConfig.PacketReceiveChanLimit)
	ServerConfig.ConnectTimeOut = beego.AppConfig.DefaultInt64("server::ConnectTimeOut", ServerConfig.ConnectTimeOut)
	ServerConfig.HeatbeatCheck = beego.AppConfig.DefaultBool("server::HeatbeatCheck", ServerConfig.HeatbeatCheck)
	ServerConfig.HeatbeatCheckSpec = beego.AppConfig.DefaultInt("server::HeatbeatCheckSpec", ServerConfig.HeatbeatCheckSpec)
}
