package utils

import (
	"net"
	"net/http"
)

func GetClientIP(r *http.Request) string {
	ipAddress, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ipAddress
}
