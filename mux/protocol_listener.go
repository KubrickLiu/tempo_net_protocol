package mux

import (
	"net"
	"sync"
)

type ProtocolListener interface {
	net.Listener

	sendConnChan(net.Conn)
	closeDoneChan()
}

type protocolListenerImpl struct {
	net.Listener
	connChan chan net.Conn
	doneChan chan struct{}
	doneOnce sync.Once
}

func newProtocolListener(l net.Listener, bufLen int) ProtocolListener {
	nl := &protocolListenerImpl{
		Listener: l,
		connChan: make(chan net.Conn, bufLen),
		doneChan: make(chan struct{}),
	}
	return nl
}

func (l *protocolListenerImpl) Accept() (net.Conn, error) {
	select {
	case c, ok := <-l.connChan:
		if !ok {
			return nil, ErrListenerClosed
		}
		return c, nil
	case <-l.doneChan:
		close(l.connChan)
		return nil, ErrListenerClosed
	}
}

func (l *protocolListenerImpl) sendConnChan(conn net.Conn) {
	l.connChan <- conn
}

func (l *protocolListenerImpl) closeDoneChan() {
	l.doneOnce.Do(func() {
		close(l.doneChan)
	})
}
