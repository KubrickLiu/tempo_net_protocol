package mux

import (
	"io"
)

type ProtocolMatcher interface {
	AppendMatchFunc(...MatchFunc)
	MatchAny(reader io.Reader) bool
	GetProtocolListener() ProtocolListener
}

type protocolMatcherImpl struct {
	protocolListener ProtocolListener
	matchers         []MatchFunc
}

func newProtocolMatcher(protocolListener ProtocolListener) ProtocolMatcher {
	pm := &protocolMatcherImpl{
		protocolListener: protocolListener,
		matchers:         make([]MatchFunc, 0),
	}
	return pm
}

func (pm *protocolMatcherImpl) AppendMatchFunc(functions ...MatchFunc) {
	for _, f := range functions {
		if f != nil {
			pm.matchers = append(pm.matchers, f)
		}
	}
}

func (pm *protocolMatcherImpl) MatchAny(reader io.Reader) bool {
	for _, match := range pm.matchers {
		if match(reader) {
			return true
		}
	}
	return false
}

func (pm *protocolMatcherImpl) GetProtocolListener() ProtocolListener {
	return pm.protocolListener
}
