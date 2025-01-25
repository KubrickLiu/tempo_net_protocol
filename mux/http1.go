package mux

import (
	"bufio"
	"io"
	"net/http"
	"strings"
)

const maxHTTPRead = 4096

// HTTP1 parses the first line or upto 4096 bytes of the request to see if
// the connection contains an HTTP request.
func HTTP1() MatchFunc {
	return func(reader io.Reader) bool {
		br := bufio.NewReader(&io.LimitedReader{R: reader, N: maxHTTPRead})
		l, part, err := br.ReadLine()
		if err != nil || part {
			return false
		}

		_, _, proto, ok := parseRequestLine(string(l))
		if !ok {
			return false
		}

		v, _, ok := http.ParseHTTPVersion(proto)
		return ok && v == 1
	}
}

// grabbed from net/http.
func parseRequestLine(line string) (method, uri, proto string, ok bool) {
	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}
