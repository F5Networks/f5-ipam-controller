package utils

import (
	"encoding/base64"
	"math/rand"
	"net"
	"strings"
)

func IsIPV4Addr(ipAddr string) bool {
	if !IsIPAddr(ipAddr) {
		return false
	}

	// presence of ":" indicates it is an IPV6
	if strings.Contains(ipAddr, ":") {
		return false
	}

	return true
}

func IsIPV6Addr(ipAddr string) bool {
	if !IsIPAddr(ipAddr) {
		return false
	}

	// presence of "." indicates it is an IPV4
	if strings.Contains(ipAddr, ".") {
		return false
	}

	return true
}

func IsIPAddr(ipAddr string) bool {
	if ipAddr == "" {
		return false
	}
	if net.ParseIP(ipAddr) == nil {
		return false
	}

	return true
}
func RandomString(len int) string {
	if len > 0 {
		buf := make([]byte, len)
		rand.Read(buf)
		str := base64.StdEncoding.EncodeToString(buf)
		// Trim to the required length
		return str[:len]
	} else {
		return ""
	}
}
