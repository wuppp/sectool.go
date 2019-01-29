package common

import (
	"net"
)

func GetIPByHost(host string) ([]string, error) {
	var ipList []string
	ips, err := net.LookupIP(host)
	if err != nil {
		// fmt.Fprintf(os.Stderr, "Could not get IPs: %v\n", err)
		return ipList, nil
	}
	for _, ip := range ips {
		ipList = append(ipList, ip.String())
	}
	return ipList, nil
}
