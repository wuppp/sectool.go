package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"sectool.go/common"
)

var (
	wg sync.WaitGroup
	ch chan bool

	f *os.File

	host       string
	port       string
	timeout    int
	threads    int
	outputFile string
)

func main() {

	options := common.PublicOptions
	flag.Parse()

	host = *options.Host
	port = *options.Port
	timeout = *options.Timeout
	threads = *options.Threads
	outputFile = *options.OutputFile

	ch = make(chan bool, threads)

	if host == "" || port == "" {
		flag.Usage()
		os.Exit(0)
	}


	if outputFile != "" {
		var err error
		f, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		common.CheckError(err)
		defer f.Close()
	}

	common.PrintInfo()

	ipList, _ := common.ParseIP(host)
	portList, _ := common.ParsePort(port)

	startTime := time.Now()
	for host := range ipList {
		for port := range portList {
			//println(fmt.Sprintf("%s:%d", host, port))
			ch <- true
			wg.Add(1)
			go scan(host, port)
		}
	}
	wg.Wait()
	scanDuration := time.Since(startTime)
	fmt.Printf("scan finished in %v", scanDuration)

}

func scan(host string, port int) {
	if isOpen(host, port) {
		fmt.Printf("%-21s open\n", fmt.Sprintf("%s:%d", host, port))
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

