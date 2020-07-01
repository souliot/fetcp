package fetcp

type SrvOption interface {
	apply(*srvOption)
}

type srvOption struct {
	Port                   int   `json:"port"`
	PacketSendChanLimit    int   `json:"packetSendChanLimit"`
	PacketReceiveChanLimit int   `json:"packetReceiveChanLimit"`
	ConnectTimeOut         int64 `json:"connectTimeOut"`
	HeatbeatCheck          bool  `json:"heatbeatCheck"`
	HeatbeatCheckSpec      int   `json:"heatbeatCheckSpec"`
}

var (
	DefaultServerConfig = &srvOption{
		Port:                   9000,
		PacketSendChanLimit:    4096,
		PacketReceiveChanLimit: 4096,
		ConnectTimeOut:         300,
		HeatbeatCheck:          true,
		HeatbeatCheckSpec:      5,
	}
)

type funcSrvOption struct {
	f func(*srvOption)
}

func (fso *funcSrvOption) apply(so *srvOption) {
	fso.f(so)
}

func newFuncSrvOption(f func(*srvOption)) *funcSrvOption {
	return &funcSrvOption{
		f: f,
	}
}

func WithPort(d int) SrvOption {
	return newFuncSrvOption(func(o *srvOption) {
		o.Port = d
	})
}

func WithPacketSendChanLimit(d int) SrvOption {
	return newFuncSrvOption(func(o *srvOption) {
		o.PacketSendChanLimit = d
	})
}

func WithPacketReceiveChanLimit(d int) SrvOption {
	return newFuncSrvOption(func(o *srvOption) {
		o.PacketReceiveChanLimit = d
	})
}

func WithConnectTimeOut(d int64) SrvOption {
	return newFuncSrvOption(func(o *srvOption) {
		o.ConnectTimeOut = d
	})
}

func WithHeatbeatCheck(d bool) SrvOption {
	return newFuncSrvOption(func(o *srvOption) {
		o.HeatbeatCheck = d
	})
}

func WithHeatbeatCheckSpec(d int) SrvOption {
	return newFuncSrvOption(func(o *srvOption) {
		o.HeatbeatCheckSpec = d
	})
}
