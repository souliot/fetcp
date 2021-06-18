package fetcp

import (
	"errors"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Error type
var (
	ErrConnClosing   = errors.New("use of closed network connection")
	ErrWriteBlocking = errors.New("write packet was blocking")
	ErrReadBlocking  = errors.New("read packet was blocking")
)

// Conn exposes a set of callbacks for the various events that occur on a connection
type Conn struct {
	srv                 *Server
	conn                *net.TCPConn  // the raw connection
	extraData           interface{}   // to save extra data
	closeOnce           sync.Once     // close the conn, once, per instance
	closeFlag           int32         // close flag
	closeChan           chan struct{} // close chanel
	packetSendChan      chan Packet   // packet send chanel
	packetReceiveChan   chan Packet   // packeet receive chanel
	HeartBeatStatus     bool          // HeartBeat status
	KeepAlive           int64         // keepAlive time
	LastTimeOfHeartBeat int64         // last HeartBeat time
}

// ConnCallback is an interface of methods that are used as callbacks on a connection
type ConnCallback interface {
	// OnConnect is called when the connection was accepted,
	// If the return value of false is closed
	OnConnect(*Conn) bool

	// OnMessage is called when the connection receives a packet,
	// If the return value of false is closed
	OnMessage(*Conn, Packet) bool

	// OnClose is called when the connection closed
	OnClose(*Conn)
}

// newConn returns a wrapper of raw conn
func newConn(conn *net.TCPConn, srv *Server) *Conn {
	return &Conn{
		srv:                 srv,
		conn:                conn,
		closeChan:           make(chan struct{}),
		packetSendChan:      make(chan Packet, srv.config.PacketSendChanLimit),
		packetReceiveChan:   make(chan Packet, srv.config.PacketReceiveChanLimit),
		HeartBeatStatus:     srv.config.HeartBeatCheck,
		KeepAlive:           srv.config.ConnectTimeOut,
		LastTimeOfHeartBeat: time.Now().Unix(),
	}
}

// GetExtraData gets the extra data from the Conn
func (c *Conn) GetExtraData() interface{} {
	return c.extraData
}

// PutExtraData puts the extra data with the Conn
func (c *Conn) PutExtraData(data interface{}) {
	c.extraData = data
}

// GetRawConn returns the raw net.TCPConn from the Conn
func (c *Conn) GetRawConn() *net.TCPConn {
	return c.conn
}

func (c *Conn) StartHeartBeatTimeOutCheck() {
	if !c.HeartBeatStatus {
		return
	}
	timer := time.NewTicker(time.Duration(c.srv.config.HeartBeatCheckSpec) * time.Second)
	if !c.IsClosed() {
		go func() {
			for _ = range timer.C {
				status := c.HeartBeatStatus
				lastTimeOfHeartBeat := c.LastTimeOfHeartBeat
				c.HeartBeatTimeOutCheck(lastTimeOfHeartBeat)
				if !status {
					timer.Stop()
				}
			}
		}()
	}
}

// Close closes the connection
func (c *Conn) Close() {
	c.closeOnce.Do(func() {
		c.HeartBeatStatus = false
		atomic.StoreInt32(&c.closeFlag, 1)
		close(c.closeChan)
		close(c.packetSendChan)
		close(c.packetReceiveChan)
		c.conn.Close()
		c.srv.DelConn(c)
		c.srv.callback.OnClose(c)
	})
}

// IsClosed indicates whether or not the connection is closed
func (c *Conn) IsClosed() bool {
	return atomic.LoadInt32(&c.closeFlag) == 1
}

// AsyncWritePacket async writes a packet, this method will never block
func (c *Conn) AsyncWritePacket(p Packet, timeout time.Duration) (err error) {
	if c.IsClosed() {
		return ErrConnClosing
	}

	defer func() {
		if e := recover(); e != nil {
			err = ErrConnClosing
		}
	}()

	if timeout == 0 {
		select {
		case c.packetSendChan <- p:
			return nil

		default:
			return ErrWriteBlocking
		}

	} else {
		select {
		case c.packetSendChan <- p:
			return nil

		case <-c.closeChan:
			return ErrConnClosing

		case <-time.After(timeout):
			return ErrWriteBlocking
		}
	}
}

// Do it
func (c *Conn) Do() {
	if !c.srv.callback.OnConnect(c) {
		return
	}
	c.srv.AddConn(c)
	asyncDo(c.StartHeartBeatTimeOutCheck, c.srv.waitGroup)
	asyncDo(c.handleLoop, c.srv.waitGroup)
	asyncDo(c.readLoop, c.srv.waitGroup)
	asyncDo(c.writeLoop, c.srv.waitGroup)
}

func (c *Conn) readLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		default:
		}

		p, err := c.srv.protocol.ReadPacket(c)
		if err != nil {
			if err == io.EOF {
				return
			}
			continue
		}
		c.UpdateLastTimeOfHeartBeat()

		c.packetReceiveChan <- p
	}
}

func (c *Conn) UpdateLastTimeOfHeartBeat() {
	c.LastTimeOfHeartBeat = time.Now().Unix()
	return
}

func (c *Conn) writeLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetSendChan:
			if c.IsClosed() {
				return
			}
			if _, err := c.conn.Write(p.Serialize()); err != nil {
				continue
			}
			c.UpdateLastTimeOfHeartBeat()
		}
	}
}

func (c *Conn) handleLoop() {
	defer func() {
		recover()
		c.Close()
	}()

	for {
		select {
		case <-c.srv.exitChan:
			return

		case <-c.closeChan:
			return

		case p := <-c.packetReceiveChan:
			if c.IsClosed() {
				return
			}
			if !c.srv.callback.OnMessage(c, p) {
				return
			}

			c.UpdateLastTimeOfHeartBeat()
		}
	}
}

func (c *Conn) HeartBeatTimeOutCheck(lastTimeOfHeartBeat int64) {
	if time.Now().Unix()-lastTimeOfHeartBeat >= c.KeepAlive {
		if !c.IsClosed() {
			c.Close()
		}
	}
}

func (c *Conn) SetKeepAlive(l int64) {
	c.KeepAlive = l
}

func (c *Conn) SetHeartBeatStatus(b bool) {
	c.HeartBeatStatus = b
}

func asyncDo(fn func(), wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		fn()
		wg.Done()
	}()
}
