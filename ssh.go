package main

import (
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sectool.go/common"

	"golang.org/x/crypto/ssh"
)

var (
	wg sync.WaitGroup
	ch chan bool

	file         string
	f            *os.File
	command      string
	user         string
	pwd          string
	userListFile string
	pwdListFile  string
	userList     []string
	pwdList      []string

	host       string
	port       string
	timeout    int
	threads    int
	outputFile string
)

func main() {

	options := common.PublicOptions
	flag.StringVar(&user, "user", "root", "username")
	flag.StringVar(&pwd, "pwd", "", "password")
	flag.StringVar(&userListFile, "uF", "", "username file path")
	flag.StringVar(&pwdListFile, "pF", "", "password file path")
	flag.StringVar(&command, "command", "id", "password file path")
	flag.Parse()

	host = *options.Host
	port = *options.Port
	timeout = *options.Timeout
	threads = *options.Threads
	outputFile = *options.OutputFile
	file = *options.File

	ch = make(chan bool, threads)

	if (host == "" || port == "") && file == "" || (user == "" && userListFile == "") || (pwd == "" && pwdListFile == "") {
		flag.Usage()
		os.Exit(0)
	}

	scanList := []string{}

	if userListFile != "" {
		userList, _ = common.ReadFileLines(userListFile)
	} else if user != "" {
		userList = append(userList, user)
	}

	if pwdListFile != "" {
		pwdList, _ = common.ReadFileLines(pwdListFile)
	} else if pwd != "" {
		pwdList = append(pwdList, pwd)
	}

	ipList, _ := common.ParseIP(host)
	portList, _ := common.ParsePort(port)

	if len(ipList) != 0 && len(portList) != 0 {
		for _, host := range ipList {
			for _, port := range portList {
				scanHost := fmt.Sprintf("%s:%d", host, port)
				scanList = append(scanList, scanHost)
			}
		}
	} else if file != "" {
		lines, err := common.ReadFileLines(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if len(portList) != 0 {
			for _, line := range lines {
				line = strings.Trim(line, " ")
				h := line
				if strings.Contains(line, ":") {
					hostPort := strings.Split(line, ":")
					h = hostPort[0]
				}
				for _, p := range portList {
					scanHost := fmt.Sprintf("%s:%d", h, p)
					scanList = append(scanList, scanHost)
				}
			}
		} else {
			for _, line := range lines {
				line = strings.Trim(line, " ")
				h := line
				p := 22
				if strings.Contains(line, ":") {
					hostPort := strings.Split(line, ":")
					h = hostPort[0]
					p, _ = strconv.Atoi(hostPort[1])
				}

				scanHost := fmt.Sprintf("%s:%d", h, p)
				scanList = append(scanList, scanHost)
			}
		}
	}

	if outputFile != "" {
		var err error
		f, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		common.CheckError(err)
		defer f.Close()
	}

	common.PrintInfo(scanList)

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

func scan(ip string, port int) {
	defer func() {
		<-ch
		wg.Done()
	}()

	for _, username := range userList {
		for _, password := range pwdList {
			if isLogin, client, err := sshLogin(ip, port, username, password); isLogin {

				var line = fmt.Sprintf("%s:%d %s %s\n", ip, port, username, password)
				f.WriteString(line)
				fmt.Printf("\033[0;32m[log] %s:%d %s %s\033[0m\n", ip, port, username, password)

				output, err := sshExec(client, command)
				if err != nil {
					fmt.Println("Failed to exec: " + err.Error())
				}
				fmt.Printf("\033[0;32m[cmd] %s:%d \n%s\033[0m", ip, port, output)
				return
			} else {
				if strings.HasPrefix(err.Error(), "ssh: ") {
					fmt.Printf("\033[0;33m[err] %s:%d %s\033[0m\n", ip, port, err)
					// fmt.Printf("[err] %s:%d %s\n", ip, port, err)
				}
			}
		}
	}
}

func sshLogin(ip string, port int, username string, password string) (bool, *ssh.Client, error) {

	clientConfig := &ssh.ClientConfig{
		Timeout: time.Duration(timeout) * time.Second,
		User:    username,
		Auth:    []ssh.AuthMethod{ssh.Password(password)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		}}

	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", ip, port), clientConfig)
	if err != nil {
		//fmt.Println("Failed to dial: " + err.Error())
		return false, nil, err
	}
	return true, conn, nil
}

func sshExec(client *ssh.Client, cmd string) (string, error) {

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	defer client.Close()

	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = os.Stderr
	session.Run(cmd)

	return stdoutBuf.String(), nil
}
