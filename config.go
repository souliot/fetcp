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
}

var (
	ServerConfig *SrvConfig
)

func init() {
	beego.LoadAppConfig("ini", "./system.ini")
	ServerConfig = &SrvConfig{
		ServerName:             "MyServer",
		Port:                   9000,
		PacketSendChanLimit:    4096,
		PacketReceiveChanLimit: 4096,
		ConnectTimeOut:         300,
	}
	ReloadConfig()
}

func ReloadConfig() {
	beego.LoadAppConfig("ini", "./system.ini")
	ServerConfig.ServerName = beego.AppConfig.String("local::ServerName")
	ServerConfig.Port, _ = beego.AppConfig.Int("local::Port")
	ServerConfig.PacketSendChanLimit, _ = beego.AppConfig.Int("server::PacketSendChanLimit")
	ServerConfig.PacketReceiveChanLimit, _ = beego.AppConfig.Int("server::PacketReceiveChanLimit")
	ServerConfig.ConnectTimeOut, _ = beego.AppConfig.Int64("server::ConnectTimeOut")
}
