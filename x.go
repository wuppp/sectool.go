package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

func main() {
	// s, _ := ParseIP("127.0.0.1/30,127.0.0.1-5,172.16.1.1")
	// fmt.Println(s)
	// s, _ := ParseIP("172.16.10.1-20.254")
	// fmt.Println(s)

	resp, err := http.Get("http://116.196.107.123/wwwroot.zip")
	if err != nil {
		// handle error
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		// handle error
	}

	fmt.Println(string(body))
}

func ParseIP(ipString string) ([]string, error) {
	ipList := []string{}

	pair := strings.Split(ipString, ",")
	for _, item := range pair {

		if net.ParseIP(item) != nil {
			ipList = append(ipList, item)
		} else if ip, network, err := net.ParseCIDR(item); err == nil {
			s := []string{}
			for ip := ip.Mask(network.Mask); network.Contains(ip); increaseIP(ip) {
				s = append(s, ip.String())
			}
			for _, ip := range s[1 : len(s)-1] {
				ipList = append(ipList, ip)
			}
		} else if strings.Contains(item, "-") {
			splitIP := strings.SplitN(item, "-", 2)
			ip := net.ParseIP(splitIP[0])
			endIP := net.ParseIP(splitIP[1])
			if endIP != nil {
				if !isStartingIPLower(ip, endIP) {
					return ipList, fmt.Errorf("%s is greater than %s", ip.String(), endIP.String())
				}
				ipList = append(ipList, ip.String())
				for !ip.Equal(endIP) {
					increaseIP(ip)
					ipList = append(ipList, ip.String())
				}
			} else {
				ipOct := strings.SplitN(ip.String(), ".", 4)
				endIP := net.ParseIP(ipOct[0] + "." + ipOct[1] + "." + ipOct[2] + "." + splitIP[1])
				if endIP != nil {
					if !isStartingIPLower(ip, endIP) {
						return ipList, fmt.Errorf("%s is greater than %s", ip.String(), endIP.String())
					}
					ipList = append(ipList, ip.String())
					for !ip.Equal(endIP) {
						increaseIP(ip)
						ipList = append(ipList, ip.String())
					}
				} else {
					return ipList, fmt.Errorf("%s is not an IP Address or CIDR Network", item)
				}
			}
		} else {
			return ipList, fmt.Errorf("%s is not an IP Address or CIDR Network", item)
		}
	}
	return ipList, nil
}

// increases an IP by a single address.
func increaseIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func isStartingIPLower(start, end net.IP) bool {
	if len(start) != len(end) {
		return false
	}
	for i := range start {
		if start[i] > end[i] {
			return false
		}
	}
	return true
}
