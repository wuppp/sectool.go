package main

import (
	. "common"
	"fmt"
	"golang.org/x/crypto/ssh"
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
	file         string
	outputFile   string
	goroutineNum int
	f            *os.File
)

func main() {

	goroutineNum = 10
	//限制goroutine数量
	ch = make(chan bool, goroutineNum)

	//flag.StringVar(&host, "host", "", "scan host. format: 127.0.0.1 | 192.168.1.1/24 | 192.168.1.1-5")
	//flag.StringVar(&port, "port", "", "scan port. format: 1-65535 | 21,22,25 | 8080")
	//flag.IntVar(&timeout, "timeout", 2, "http connect timeout")
	//flag.StringVar(&file, "f", "", "123")
	//flag.BoolVar(&verbose, "v", false, "show verbose")
	//flag.IntVar(&goroutineNum, "t", 5000, "scan thread number. Default 5000")
	//flag.StringVar(&outputFile, "o", "", "save result file")
	//flag.Parse()
	//scan("104.238.160.159", 22)
	lines, _ := ReadFileLines("./sectool.go/portscan/ssh.txt")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			splitLine := strings.SplitN(line, ":", 2)
			ip := splitLine[0]
			port, _ := strconv.Atoi(splitLine[1])

			userList := []string{"root"}
			passwordList := []string{"123456", "118.193.151.15"}

			for _, username := range userList {
				for _, password := range passwordList {
					ch <- true
					wg.Add(1)
					go scan(ip, port, username, password)
				}
			}
		}
	}
}

func scan(ip string, port int, username string, password string) {
	if sshLogin(ip, port, username, password) {
		fmt.Printf("[+] %s:%d", ip, port)
	}

}

func sshLogin(ip string, port int, username string, password string) bool {

	defer func() {
		<-ch
		wg.Done()
	}()

	clientConfig := &ssh.ClientConfig{
		Timeout: time.Duration(2) * time.Second,
		User:    username,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}}

	_, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), clientConfig)
	if err != nil {
		//panic("Failed to dial: " + err.Error())
		//fmt.Println("Failed to dial: " + err.Error())
		return false
	}

	return true
}
