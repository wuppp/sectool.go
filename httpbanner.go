package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"sectool.go/common"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

var (
	wg sync.WaitGroup
	ch chan bool

	file       string
	f          *os.File
	reqHost    string
	path       string
	redirect   bool
	grepString string
	code       int
	result     []HttpInfo

	host       string
	port       string
	timeout    int
	threads    int
	outputFile string
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

	options := common.PublicOptions
	flag.StringVar(&path, "path", "/", "request url path. /phpinfo.php | /index.html")
	flag.BoolVar(&redirect, "redirect", false, "follow 30x redirect")
	flag.Var(&reqHeaders, "H", "request headers. exmaple: -H User-Agent: xx -H Referer: xx")
	flag.StringVar(&grepString, "grep", "", "response body grep string. -grep phpinfo")
	flag.IntVar(&code, "code", 0, "response status code grep. -code 200")
	flag.Parse()

	host = *options.Host
	port = *options.Port
	timeout = *options.Timeout
	threads = *options.Threads
	outputFile = *options.OutputFile
	file = *options.File

	ch = make(chan bool, threads)

	if (host == "" || port == "") && file == "" {
		flag.Usage()
		os.Exit(0)
	}

	ipList, _ := common.ParseIP(host)
	portList, _ := common.ParsePort(port)
	scanList := []string{}

	// 处理请求头
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
		lines = common.ParseLines(lines)

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
		common.CheckError(err)
		defer f.Close()
	}

	fmt.Printf("host: %s\n", host)
	fmt.Printf("port: %s\n", port)
	fmt.Printf("path: %s\n", path)
	fmt.Printf("file: %s\n", file)
	fmt.Printf("redirect: %t\n", redirect)
	fmt.Printf("grep: %s\n", grepString)
	fmt.Printf("code: %d\n", code)
	fmt.Printf("headers:\n")

	for k, v := range headers {
		fmt.Printf("    %s: %s\n", k, v)
	}
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
		go scan(url)
	}
	wg.Wait()
	scanDuration := time.Since(startTime)
	fmt.Printf("scan finished in %v", scanDuration)
}

func scan(url string) {
	fetch(url)
	<-ch
	wg.Done()
}

func determineEncoding(r *bufio.Reader) encoding.Encoding {
	b, err := r.Peek(1024)
	if err != nil {
		// log.Error("get code error")
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(b, "")
	return e
}

func fetch(url string) {

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

	// 获取编码
	reader := bufio.NewReader(resp.Body)
	e := determineEncoding(reader)
	utf8Reader := transform.NewReader(reader, e.NewDecoder())

	// 获取标题
	body, err := ioutil.ReadAll(utf8Reader)
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

	// 响应头提取字段
	info.Server = resp.Header.Get("Server")
	info.ContentLength = resp.Header.Get("Content-Length")
	info.XPoweredBy = resp.Header.Get("X-Powered-By")
	pair := strings.SplitN(resp.Header.Get("Content-Type"), ";", 2)
	if len(pair) == 2 {
		info.ContentType = pair[0]
	}
	result = append(result, *info)

	statusCode := strconv.Itoa(info.StatusCode)

	if strings.Contains(respBody, grepString) && (code == 0 || strings.HasPrefix(statusCode, strconv.Itoa(code))) {
		var line = fmt.Sprintf("%-5d %-6s %-16s %-55s %-20s %-50s %s\n", info.StatusCode, info.ContentLength, info.ContentType, info.Server, info.XPoweredBy, info.Url, info.Title)
		if runtime.GOOS == "windows" {
			fmt.Printf(line)
		} else {
			if strings.HasPrefix(statusCode, "2") {
				fmt.Printf("\033[0;32m%s\033[0m", line)
			} else if strings.HasPrefix(statusCode, "3") {
				fmt.Printf("\033[0;35m%s\033[0m", line)
			} else if strings.HasPrefix(statusCode, "4") {
				fmt.Printf("\033[0;33m%s\033[0m", line)
			} else if strings.HasPrefix(statusCode, "5") {
				fmt.Printf("\033[0;31m%s\033[0m", line)
			} else {
				fmt.Printf(line)
			}
		}
		f.WriteString(line)
	}

}
