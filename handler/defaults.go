package handler

import (
	"fmt"
	"net"
)

const (
	DefaultPushPort     = 10087
	DefaultTCPPort      = 10086
	DefaultUDPPort      = 10088
	DefaultMQTTPort     = 1883
	DefaultSeleniumPort = 20165
	DefaultHTTPPort     = 8080
	DefaultWebsitePort  = 8000
)

func normalizePushAddress(address string) string {
	switch address {
	case "", "self":
		return fmt.Sprintf("localhost:%d", DefaultPushPort)
	case "all", "ll", "1":
		return fmt.Sprintf("255.255.255.255:%d", DefaultPushPort)
	default:
		return fmt.Sprintf("%s:%d", address, DefaultPushPort)
	}
}

func resolveBroadcastUDPAddr(address string) (*net.UDPAddr, error) {
	return net.ResolveUDPAddr("udp", normalizePushAddress(address))
}
