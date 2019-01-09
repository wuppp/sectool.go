package main

import (
	"common"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	wg           sync.WaitGroup
	ch           chan bool
	host         string
	reqHost      string
	port         string
	path         string
	file         string
	timeout      int
	redirect     bool
	outputFile   string
	goroutineNum int
	result       []HttpInfo
	f            *os.File
)

var headers = map[string]string{
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36",
}

var reqHeaders arrayFlags

type arrayFlags []string

type HttpInfo struct {
	StatusCode    int    `json:"status_code"`
	Url           string `json:"url"`
	Title         string `json:"title"`
	Server        string `json:"server"`
	ContentLength string `json:"length"`
	ContentType   string `json:"type"`
	XPoweredBy    string `json:xpoweredby`
}

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {

	flag.StringVar(&host, "host", "", "scan hosts")
	flag.IntVar(&timeout, "timeout", 5, "http connect timeout")
	flag.StringVar(&port, "p", "", "scan port")
	flag.StringVar(&file, "f", "", "load external file")
	flag.IntVar(&goroutineNum, "t", 500, "scan thread number. Default 500")
	flag.StringVar(&outputFile, "o", "", "save result file")
	flag.StringVar(&path, "path", "/", "request path example: /admin")
	flag.BoolVar(&redirect, "redirect", false, "follow 30x redirect")
	flag.Var(&reqHeaders, "H", "request headers. exmaple: -H User-Agent: xx -H Referer: xx")
	flag.Parse()

	// prepare request headers
	for _, line := range reqHeaders {
		pair := strings.SplitN(line, ":", 2)
		if len(pair) == 2 {
			k, v := pair[0], strings.Trim(pair[1], " ")
			if strings.ToLower(k) == "host" {
				reqHost = v
			}
			headers[k] = v
		}
	}

	//限制goroutine数量
	ch = make(chan bool, goroutineNum)

	if host == "" && file == "" {
		flag.Usage()
		os.Exit(0)
	}

	scanList := []string{}
	ipList, _ := common.ParseIP(host)
	portList, _ := common.ParsePort(port)

	if len(ipList) != 0 && len(portList) != 0 {
		for _, host := range ipList {
			for _, port := range portList {
				scanHost := fmt.Sprintf("%s:%d", host, port)
				scanList = append(scanList, scanHost)
			}
		}
	}

	if file != "" {
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
				p := 80
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
		checkError(err)
		defer f.Close()
	}

	fmt.Printf("host: %s\n", host)
	fmt.Printf("port: %s\n", port)
	fmt.Printf("path: %s\n", path)
	fmt.Printf("file: %s\n", file)
	fmt.Printf("headers:\n")

	for k, v := range headers {
		fmt.Printf("    %s: %s\n", k, v)
	}
	fmt.Printf("\n")
	fmt.Printf("Number of scans: %d\n", len(scanList))

	startTime := time.Now()
	for _, line := range scanList {
		ch <- true
		wg.Add(1)

		pair := strings.SplitN(line, ":", 2)
		host := pair[0]
		port, _ := strconv.Atoi(pair[1])
		url := fmt.Sprintf("http://%s:%d%s", host, port, path)
		if port == 443 {
			url = fmt.Sprintf("https://%s%s", host, path)
		}
		go fetch(url)
	}
	wg.Wait()
	scanDuration := time.Since(startTime)
	fmt.Printf("scan finished in %v", scanDuration)

}

func fetch(url string) {

	defer func() {
		<-ch
		wg.Done()
	}()
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
	}
	if !redirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// log.Println("")
	}

	req.Host = reqHost
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		// log.Println("http.Get:", err.Error())
		return
	}
	defer resp.Body.Close()

	info := &HttpInfo{}
	info.Url = url
	info.StatusCode = resp.StatusCode

	// 获取标题
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// log.Println("ioutil.ReadAll", err.Error())
		return
	}
	respBody := string(body)
	r := regexp.MustCompile(`(?i)<title>\s?(.*?)\s?</title>`)
	m := r.FindStringSubmatch(respBody)
	if len(m) == 2 {
		info.Title = m[1]
	}

	// 获取响应头Server字段
	info.Server = resp.Header.Get("Server")
	info.ContentLength = resp.Header.Get("Content-Length")
	info.XPoweredBy = resp.Header.Get("X-Powered-By")
	pair := strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)
	if len(pair) == 2 {
		info.ContentType = pair[0]
	}

	result = append(result, *info)

	var line = fmt.Sprintf("%-5d %-6s %-16s %-40s %-20s %-50s %s\n", info.StatusCode, info.ContentLength, info.ContentType, info.Server, info.XPoweredBy, info.Url, info.Title)
	fmt.Printf(line)
	f.WriteString(line)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
