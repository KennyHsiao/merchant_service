package merchantsService

import (
	"net"
	"strings"
)

func IPChecker(myip string, whitelist string) bool {

	// 47.75.121.127 是反向代理IP
	if myip == "localhost" || myip == "127.0.0.1" || myip == "0:0:0:0:0:0:0:1" || myip == "47.75.121.127" {
		return true
	}

	if whitelist == "" {
		return false
	}
	for _, ip := range strings.Split(whitelist, ",") {
		if !strings.Contains(ip, "/") {
			ip = ip + "/32"
		}
		_, ipnetA, _ := net.ParseCIDR(ip)
		if ipnetA == nil {
			continue
		}
		ipB := net.ParseIP(myip)

		if ipnetA.Contains(ipB) {
			return true
		}
	}
	return false
}
