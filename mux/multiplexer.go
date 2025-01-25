package mux

import (
	"io"
	"net"
	"strings"
	"sync"
)

type MatchFunc func(reader io.Reader) bool

type ErrorHandler func(error) bool

type Multiplexer interface {
	io.Closer

	Serve() error
	RegisterNewProtocolListener(...MatchFunc) net.Listener
	ErrorHandler(ErrorHandler)
}

type multiplexer struct {
	root       net.Listener
	bufLen     int
	doneChan   chan struct{}
	doneOnce   sync.Once
	matchers   []ProtocolMatcher
	errHandler ErrorHandler
}

func NewMux(l net.Listener, bufLen int) Multiplexer {
	mux := &multiplexer{
		root:       l,
		bufLen:     bufLen,
		doneChan:   make(chan struct{}),
		matchers:   make([]ProtocolMatcher, 0),
		errHandler: func(_ error) bool { return true },
	}
	return mux
}

func (mux *multiplexer) Serve() error {
	var wg sync.WaitGroup
	var closeOnce sync.Once

	closeSubListener := func() {
		closeOnce.Do(func() {
			wg.Wait()

			for _, matcher := range mux.matchers {
				pl := matcher.GetProtocolListener()
				pl.closeDoneChan()
			}
		})
	}

	defer func() {
		closeSubListener()
	}()

	for {
		select {
		case <-mux.doneChan:
			closeSubListener()
			return ErrServerClosed
		default:
			conn, err := mux.root.Accept()
			if err != nil {
				if strings.EqualFold(err.Error(), "use of closed network connection") {
					closeSubListener()
					return ErrServerClosed
				}

				if !mux.handleError(err) {
					return err
				}
				continue
			}

			wg.Add(1)
			mux.serve(conn, &wg)
		}
	}
}

func (mux *multiplexer) serve(conn net.Conn, wg *sync.WaitGroup) {
	defer wg.Done()

	muxConn := newMuxConn(conn)
	for _, matcher := range mux.matchers {
		muxConn.startSniffing()
		if matcher.MatchAny(muxConn) {
			muxConn.doneSniffing()

			listener := matcher.GetProtocolListener()
			listener.sendConnChan(muxConn)
		}
	}
}

func (mux *multiplexer) handleError(err error) bool {
	return mux.errHandler(err)
}

func (mux *multiplexer) RegisterNewProtocolListener(functions ...MatchFunc) net.Listener {
	pl := newProtocolListener(mux.root, mux.bufLen)
	pm := newProtocolMatcher(pl)
	pm.AppendMatchFunc(functions...)

	mux.matchers = append(mux.matchers, pm)
	return pl
}

func (mux *multiplexer) ErrorHandler(h ErrorHandler) {
	mux.errHandler = h
}

func (mux *multiplexer) Close() error {
	mux.doneOnce.Do(func() {
		close(mux.doneChan)
		mux.root.Close()
	})
	return nil
}
