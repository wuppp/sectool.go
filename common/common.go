package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	//"sort"
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

func IsValidIPV4(ip string) bool {
	b := net.ParseIP(ip)
	if b.To4() == nil {
		return false
	}
	return true
}

func CheckPort(portString string) error {
	pair := strings.Split(portString, ",")
	for _, item := range  pair {
		if strings.Contains(item, "-") {
			portRange := strings.Split(item, "-")
			if len(portRange) != 2 {
				return fmt.Errorf("%s is invalid port range", portString)
			}
		}
	}
	return nil
}

func ParsePort(portString string) (<-chan int, error) {

	portList := make(chan int)

	go func() {
		defer close(portList)
		pair := strings.Split(portString, ",")
		for _, item := range pair {
			if strings.Contains(item, "-") {
				portRange := strings.Split(item, "-")
				if len(portRange) != 2 {
					return
					//return portList, fmt.Errorf("%s is invalid port range", portString)
				}
				start, _ := strconv.Atoi(portRange[0])
				end, _ := strconv.Atoi(portRange[1])
				for i := start; i <= end; i++ {
					portList <- i
				}
			} else {
				if item != "" {
					item, _ := strconv.Atoi(item)
					portList <- item
				}
			}
		}

	}()

	return portList, nil
}

func ParseIP(ipString string) (<-chan string, error) {
	ipList := make(chan string)

	go func() {
		defer close(ipList)
		pair := strings.Split(ipString, ",")
		for _, item := range pair {
			if net.ParseIP(item) != nil {
				ipList <- item
			} else if ip, network, err := net.ParseCIDR(item); err == nil {
				flag := true
				for ip := ip.Mask(network.Mask); network.Contains(ip); increaseIP(ip) {
					if flag {
						flag = false
						continue
					}
					ipList <- ip.String()
				}
			} else if strings.Contains(item, "-") {
				splitIP := strings.SplitN(item, "-", 2)
				ip := net.ParseIP(splitIP[0])
				endIP := net.ParseIP(splitIP[1])
				end, _ := strconv.Atoi(splitIP[1])
				if endIP != nil {
					if !isStartingIPLower(ip, endIP) {
						return
						//return fmt.Errorf("%s is greater than %s", ip.String(), endIP.String())
					}
					ipList <- ip.String()
					for !ip.Equal(endIP) {
						increaseIP(ip)
						ipList <- ip.String()
					}
				} else if end >= 1 && end <= 255 {
					ipOct := strings.SplitN(ip.String(), ".", 4)
					if len(ipOct) == 4 {
						a, _ := strconv.Atoi(ipOct[0])
						b, _ := strconv.Atoi(ipOct[1])
						c, _ := strconv.Atoi(ipOct[2])
						if (a >= 1 && a <= 255) && (b >= 1 && b <= 255) && (c >= 1 && c <= 255) {
							endIP := net.ParseIP(ipOct[0] + "." + ipOct[1] + "." + ipOct[2] + "." + splitIP[1])
							if endIP != nil {
								if !isStartingIPLower(ip, endIP) {
									//fmt.Errorf("%s is greater than %s", ip.String(), endIP.String())
									return

								}
								ipList <- ip.String()
								for !ip.Equal(endIP) {
									increaseIP(ip)
									ipList <- ip.String()
								}
							} else {
								return
								//return ipList, fmt.Errorf("%s is not an IP Address or CIDR Network", item)
							}
						}
					}
				} else {
					return
					//return ipList, fmt.Errorf("%s is not an IP Address or CIDR Network", item)
				}
			}
		}
	}()
	//
	//pair := strings.Split(ipString, ",")
	//for _, item := range pair {
	//
	//	if net.ParseIP(item) != nil {
	//		ipList = append(ipList, item)
	//	} else if ip, network, err := net.ParseCIDR(item); err == nil {
	//		s := []string{}
	//		for ip := ip.Mask(network.Mask); network.Contains(ip); increaseIP(ip) {
	//			s = append(s, ip.String())
	//		}
	//		for _, ip := range s[1 : len(s)-1] {
	//			ipList = append(ipList, ip)
	//		}
	//	} else if strings.Contains(item, "-") {
	//		splitIP := strings.SplitN(item, "-", 2)
	//		ip := net.ParseIP(splitIP[0])
	//		endIP := net.ParseIP(splitIP[1])
	//		end, _ := strconv.Atoi(splitIP[1])
	//		if endIP != nil {
	//			if !isStartingIPLower(ip, endIP) {
	//				return ipList, fmt.Errorf("%s is greater than %s", ip.String(), endIP.String())
	//			}
	//			ipList = append(ipList, ip.String())
	//			for !ip.Equal(endIP) {
	//				increaseIP(ip)
	//				ipList = append(ipList, ip.String())
	//			}
	//		} else if end >= 1 && end <= 255 {
	//			ipOct := strings.SplitN(ip.String(), ".", 4)
	//			if len(ipOct) == 4 {
	//				a, _ := strconv.Atoi(ipOct[0])
	//				b, _ := strconv.Atoi(ipOct[1])
	//				c, _ := strconv.Atoi(ipOct[2])
	//				if (a >= 1 && a <= 255) && (b >= 1 && b <= 255) && (c >= 1 && c <= 255) {
	//					endIP := net.ParseIP(ipOct[0] + "." + ipOct[1] + "." + ipOct[2] + "." + splitIP[1])
	//					if endIP != nil {
	//						if !isStartingIPLower(ip, endIP) {
	//							return ipList, fmt.Errorf("%s is greater than %s", ip.String(), endIP.String())
	//						}
	//						ipList = append(ipList, ip.String())
	//						for !ip.Equal(endIP) {
	//							increaseIP(ip)
	//							ipList = append(ipList, ip.String())
	//						}
	//					} else {
	//						return ipList, fmt.Errorf("%s is not an IP Address or CIDR Network", item)
	//					}
	//				}
	//			}
	//		}
	//	} else {
	//		return ipList, fmt.Errorf("%s is not an IP Address or CIDR Network", item)
	//	}
	//}
	return ipList, nil
}

// LinesToIPList processes a list of IP addresses or networks in CIDR format.
// Returning a list of all possible IP addresses.
//func LinesToIPList(lines []string) ([]string, error) {
//	ipList := []string{}
//	for _, line := range lines {
//		_ipList, err := ParseIP(line)
//		if err != nil {
//			return _ipList, fmt.Errorf("%s is not an IP Address", line)
//		}
//		for _, line := range _ipList {
//			ipList = append(ipList, line)
//		}
//	}
//	return ipList, nil
//}

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

//func ParseLine(line string) []string {
//	lines := make(chan string)
//
//	go func() {
//		defer close(lines)
//
//		ips, _ := ParseIP(line)
//		for ip := range ips {
//			lines <- ip
//		}
//	}()
//	ips, _ := ParseIP(line)
//
//	for _, line := range l {
//		ips, _ := ParseIP(line)
//		if len(ips) != 0 {
//			for _, ip := range ips {
//				lines = append(lines, ip)
//			}
//		} else {
//			// ips, err := GetIPByHost(line)
//			// if err != nil && len(ips) != 0 {
//			// }
//			lines = append(lines, line)
//		}
//	}
//	return lines
//}

// ReadFileLines returns all the lines in a file.
func ReadFileLines(path string) (<-chan string, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	defer file.Chdir()

	scanner := bufio.NewScanner(file)

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	line := make(chan string)

	go func() {
		defer close(line)

		for scanner.Scan() {
			if scanner.Text() == "" {
				continue
			}
			ips, _ := ParseIP(scanner.Text())
			for ip := range ips {
				line <- ip
			}
		}
	}()

	return line, nil
}

func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

