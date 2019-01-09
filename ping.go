package main

import (
	"common"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	wg           sync.WaitGroup
	ch           chan bool
	host         string
	timeout      int
	goroutineNum int
	args         []string
	outputFile   string
	f            *os.File
)

func main() {
	flag.StringVar(&host, "host", "", "scan host. format: 127.0.0.1 | 192.168.1.1/24 | 192.168.1.1-5")
	flag.IntVar(&timeout, "timeout", 2, "ping connect timeout")
	flag.IntVar(&goroutineNum, "t", 100, "scan thread number. Default 100")
	flag.StringVar(&outputFile, "o", "", "save result file")
	flag.Parse()

	ch = make(chan bool, goroutineNum)

	if host == "" {
		flag.Usage()
		os.Exit(0)
	}

	switch runtime.GOOS {
	case "darwin":
		args = []string{"-c", "1", "-W", strconv.Itoa(timeout * 1000)}
	case "windows":
		args = []string{"-n", "1", "-w", strconv.Itoa(timeout)}
	case "linux":
		args = []string{"-c", "1", "-W", strconv.Itoa(timeout)}
	default:
		args = []string{"-c", "1", "-W", strconv.Itoa(timeout)}
	}

	if outputFile != "" {
		var err error
		f, err = os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
		checkError(err)
		defer f.Close()
	}

	ipList, _ := common.ParseIP(host)

	startTime := time.Now()
	for _, ip := range ipList {
		ch <- true
		wg.Add(1)
		go scan(ip)
	}
	wg.Wait()
	scanDuration := time.Since(startTime)
	fmt.Printf("scan finished in %v", scanDuration)
}

func scan(ip string) {
	result, ms := run(ip)
	if result {
		fmt.Printf("%-15s %sms\n", ip, ms)
		f.WriteString(fmt.Sprintf("%-15s %sms\n", ip, ms))
	}
	<-ch
	wg.Done()
}

func run(ip string) (bool, string) {
	pingArgs := append(args, ip)
	cmd := exec.Command("ping", pingArgs...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// log.Printf("cmd.Run() failed with %s\n", err)
	}

	if strings.Contains(strings.ToLower(string(out)), "ttl") {
		// fmt.Printf("combined out:\n%s\n", string(out))
		r := regexp.MustCompile("[<>=]([0-9.]+)\\s?ms")
		m := r.FindStringSubmatch(string(out))
		if len(m) == 2 {
			return true, m[1]
		}
	}
	return false, ""
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
