package main

import (
	. "common"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	wg           sync.WaitGroup
	ch           chan bool
	host         string
	port         string
	timeout      int
	verbose      bool
	outputFile   string
	goroutineNum int
	f            *os.File
)

func main() {

	flag.StringVar(&host, "host", "", "scan host. format: 127.0.0.1 | 192.168.1.1/24 | 192.168.1.1-5")
	flag.StringVar(&port, "p", "", "scan port. format: 1-65535 | 21,22,25 | 8080")
	flag.IntVar(&timeout, "timeout", 2, "http connect timeout")
	flag.BoolVar(&verbose, "v", false, "show verbose")
	flag.IntVar(&goroutineNum, "t", 2000, "scan thread number. Default 2000")
	flag.StringVar(&outputFile, "o", "", "save result file")
	flag.Parse()

	//限制goroutine数量
	ch = make(chan bool, goroutineNum)

	if host == "" || port == "" {
		flag.Usage()
		os.Exit(0)
	}

	ipList, _ := ParseIP(host)
	portList, _ := ParsePort(port)

	scanList := []string{}
	for _, host := range ipList {
		for _, port := range portList {
			scanHost := fmt.Sprintf("%s:%d", host, port)
			scanList = append(scanList, scanHost)
		}
	}

	if outputFile != "" {
		var err error
		f, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		checkError(err)
		defer f.Close()
	}

	fmt.Printf("host: %s\n", host)
	fmt.Printf("port: %s\n", port)
	fmt.Printf("Number of scans: %d\n", len(scanList))

	startTime := time.Now()
	for _, line := range scanList {
		ch <- true
		wg.Add(1)

		pair := strings.SplitN(line, ":", 2)
		host := pair[0]
		port, _ := strconv.Atoi(pair[1])
		go scan(host, port)
	}
	wg.Wait()
	scanDuration := time.Since(startTime)
	fmt.Printf("scan finished in %v", scanDuration)

}

func scan(host string, port int) {
	if isOpen(host, port) {
		fmt.Printf("%s open\n", ljust(fmt.Sprintf("%s:%d", host, port), 21))
		f.WriteString(fmt.Sprintf("%s:%d\r\n", host, port))
	}
	<-ch
	wg.Done()
}

func isOpen(host string, port int) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Duration(timeout)*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
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

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
