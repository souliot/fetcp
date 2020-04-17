package fetcp

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
	DefaultServerConfig = &SrvConfig{
		ServerName:             "fetcp",
		Port:                   9000,
		PacketSendChanLimit:    4096,
		PacketReceiveChanLimit: 4096,
		ConnectTimeOut:         300,
		HeatbeatCheck:          false,
		HeatbeatCheckSpec:      5,
	}
)

func (s *SrvConfig) MergeConfig(cs ...*SrvConfig) {
	if len(cs) < 1 {
		return
	}

	sc := cs[0]
	if sc.ServerName != "" {
		s.ServerName = sc.ServerName
	}
	if sc.Port != 0 {
		s.Port = sc.Port
	}
	if sc.PacketSendChanLimit != 0 {
		s.PacketSendChanLimit = sc.PacketSendChanLimit
	}
	if sc.PacketReceiveChanLimit != 0 {
		s.PacketReceiveChanLimit = sc.PacketReceiveChanLimit
	}
	if sc.ConnectTimeOut != 0 {
		s.ConnectTimeOut = sc.ConnectTimeOut
	}
	if sc.HeatbeatCheckSpec != 0 {
		s.HeatbeatCheckSpec = sc.HeatbeatCheckSpec
	}
	s.HeatbeatCheck = sc.HeatbeatCheck

}
