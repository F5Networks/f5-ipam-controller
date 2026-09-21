package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/google/uuid"
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
		id := uuid.New()
		return id.String()[:len]
	} else {
		return ""
	}
}
func Ipv4ToPaddedString(ip string) string {
	splitIp := strings.Split(ip, ".")
	return fmt.Sprintf("%03s.%03s.%03s.%03s", splitIp[0], splitIp[1], splitIp[2], splitIp[3])
}
func PaddedStringToIPV4(paddedIp string) string {
	splitIp := strings.Split(paddedIp, ".")
	if len(splitIp) != 4 {
		return paddedIp
	}
	octets := make([]string, 4)
	for i, o := range splitIp {
		n, err := strconv.Atoi(o)
		if err != nil {
			// not a padded numeric octet; return input unchanged
			return paddedIp
		}
		octets[i] = strconv.Itoa(n) // "000" -> "0", "082" -> "82"
	}
	return strings.Join(octets, ".")
}
