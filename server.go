package fetcp

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/astaxie/beego"
)

type Server struct {
	config    *SrvConfig      // server configuration
	callback  ConnCallback    // message callbacks in connection
	protocol  Protocol        // customize packet protocol
	exitChan  chan struct{}   // notify all goroutines to shutdown
	waitGroup *sync.WaitGroup // wait for all goroutines
}

// NewServer creates a server
func NewServer(callback ConnCallback, protocol Protocol) *Server {
	return &Server{
		config:    ServerConfig,
		callback:  callback,
		protocol:  protocol,
		exitChan:  make(chan struct{}),
		waitGroup: &sync.WaitGroup{},
	}
}

func (s *Server) Server() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		beego.Error("resolve tcp addr err: ", err)
		return
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		beego.Error("listen err", err)
		return
	}
	go s.Start(listener, time.Second)
	beego.Info("应用", s.ServerName, "启动监听：", listener.Addr(), "成功")
}

// Start starts service
func (s *Server) Start(listener *net.TCPListener, acceptTimeout time.Duration) {
	s.waitGroup.Add(1)
	defer func() {
		listener.Close()
		s.waitGroup.Done()
	}()

	for {
		select {
		case <-s.exitChan:
			return

		default:
		}

		listener.SetDeadline(time.Now().Add(acceptTimeout))

		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}

		s.waitGroup.Add(1)
		go func() {
			newConn(conn, s).Do()
			s.waitGroup.Done()
		}()
	}
}

// Stop stops service
func (s *Server) Stop() {
	close(s.exitChan)
	s.waitGroup.Wait()
}
