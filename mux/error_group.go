package mux

import "errors"

var ErrListenerClosed = errors.New("mux: listener closed")

var ErrServerClosed = errors.New("mux: server closed")
