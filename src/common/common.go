package common

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func center(s string, width int) string {
	n := width - len(s)
	if n <= 0 {
		return s
	}
	half := n / 2
	if n%2 != 0 && width%2 != 0 {
		half = half + 1
	}
	return strings.Repeat(" ", half) + s + strings.Repeat(" ", (n-half))
}

func rjust(s string, width int) string {
	n := width - len(s)
	if n <= 0 {
		return s
	}
	return strings.Repeat(" ", n) + s
}

func ljust(s string, width int) string {
	n := width - len(s)
	if n <= 0 {
		return s
	}
	return s + strings.Repeat(" ", n)
}

func ParseHost(hostString string) []string {
	var hosts []string

	pair := strings.Split(hostString, ",")
	for _, item := range pair {
		if strings.Contains(item, "/") {
			_hosts, err := ParseCIDR(item)
			checkError(err)
			for _, v := range _hosts {
				hosts = append(hosts, v)
			}
		} else if strings.Contains(item, "-") {

			pair := strings.Split(item, ".")
			prefix := strings.Join(pair[0:len(pair)-1], ".")

			r := regexp.MustCompile(`(\d+)-(\d+)`)
			hostRange := r.FindStringSubmatch(item)

			if len(hostRange) == 3 {
				start, _ := strconv.Atoi(hostRange[1])
				end, _ := strconv.Atoi(hostRange[2])
				for i := start; i <= end; i++ {
					ip_str := strings.Join([]string{prefix, strconv.Itoa(i)}, ".")
					hosts = append(hosts, ip_str)
				}
			}

		} else if isValidIPV4(item) {
			hosts = append(hosts, item)
		}
	}

	return hosts
}

func isValidIPV4(ip string) bool {
	b := net.ParseIP(ip)
	if b.To4() == nil {
		return false
	}
	return true
}

func ParseCIDR(cidr string) ([]string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil

}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func ParsePort(portString string) []int {

	var ports []int

	pair := strings.Split(portString, ",")
	for _, item := range pair {
		if strings.Contains(item, "-") {
			portRange := strings.Split(item, "-")
			if len(portRange) == 2 {
				start, _ := strconv.Atoi(portRange[0])
				end, _ := strconv.Atoi(portRange[1])
				for i := start; i <= end; i++ {
					ports = append(ports, i)
				}
			}
		} else {
			item, _ := strconv.Atoi(item)
			ports = append(ports, item)
		}
	}

	sort.Ints(ports)
	return ports
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
