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
	wg sync.WaitGroup
	ch chan bool

	f    *os.File
	args []string

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
		common.CheckError(err)
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
		// fmt.Printf("cmd.Run() failed with %s\n", err)
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
