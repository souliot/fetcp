package fetcp

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type Server struct {
	config    *srvOption      // server configuration
	callback  ConnCallback    // message callbacks in connection
	protocol  Protocol        // customize packet protocol
	exitChan  chan struct{}   // notify all goroutines to shutdown
	waitGroup *sync.WaitGroup // wait for all goroutines
	conns     *sync.Map
}

func NewServer(callback ConnCallback, protocol Protocol, opts ...SrvOption) *Server {
	s := &Server{
		config:    DefaultServerConfig,
		callback:  callback,
		protocol:  protocol,
		exitChan:  make(chan struct{}),
		waitGroup: new(sync.WaitGroup),
		conns:     new(sync.Map),
	}
	for _, opt := range opts {
		opt.apply(s.config)
	}
	return s
}

func (s *Server) Server() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		fmt.Println("resolve tcp addr err: ", err)
		return
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("listen err", err)
		return
	}
	s.waitGroup.Wait()
	go s.Start(listener, time.Second)
}

func (s *Server) GetOptions() *srvOption {
	return s.config
}

func (s *Server) AddConn(c *Conn) {
	s.conns.Store(c, true)
}

func (s *Server) DelConn(c *Conn) {
	s.conns.Delete(c)
}

func (s *Server) GetConns() []*Conn {
	cs := make([]*Conn, 0)
	s.conns.Range(func(k, v interface{}) bool {
		cs = append(cs, k.(*Conn))
		return true
	})
	return cs
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
	cs := s.GetConns()
	for _, v := range cs {
		v.Close()
	}
	close(s.exitChan)
}
