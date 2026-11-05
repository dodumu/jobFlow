package utils

import (
	"net"
	"net/http"
	"os"
	"strings"
)

func GetClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	if os.Getenv("TRUST_PROXY") != "true" {
		return host
	}

	header := r.Header.Get("X-Forwarded-For")
	parts := strings.Split(header, ",")
	last := strings.TrimSpace(parts[len(parts)-1])

	if net.ParseIP(last) != nil {
		return last
	}

	return host

}
