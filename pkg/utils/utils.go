package utils

import (
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
