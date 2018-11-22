package main

import (
	. "common"
	"crypto/tls"
	"encoding/json"
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
	wg             sync.WaitGroup
	ch             chan bool
	host           string
	reqHost        string
	port           string
	path           string
	file           string
	timeout        int
	redirect       bool
	outputJSONFile string
	goroutineNum   int
	result         []HttpInfo
)

var headers = map[string]string{
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36",
}

var reqHeaders arrayFlags

type arrayFlags []string

type HttpInfo struct {
	StatusCode int    `json:"status_code"`
	Url        string `json:"url"`
	Title      string `json:"title"`
	Server     string `json:"server"`
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
	flag.StringVar(&port, "p", "80", "scan port. Default 80")
	flag.StringVar(&file, "f", "", "load external file")
	flag.IntVar(&goroutineNum, "t", 1000, "scan thread number. Default 1000")
	flag.StringVar(&outputJSONFile, "oJ", "", "save result file")
	flag.StringVar(&path, "path", "/", "save result file")
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

	if host != "" && port != "" {
		ipList, _ := ParseIP(host)
		portList, _ := ParsePort(port)
		for _, host := range ipList {
			for _, port := range portList {
				scanHost := fmt.Sprintf("%s:%d", host, port)
				scanList = append(scanList, scanHost)
			}
		}
	}

	if file != "" {
		lines, err := ReadFileLines(file)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, line := range lines {
			scanList = append(scanList, line)
		}
	}

	fmt.Printf("host: %s\n", host)
	fmt.Printf("port: %s\n", port)
	fmt.Printf("path: %s\n", path)
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

	if outputJSONFile != "" {
		saveResult(result)
	}
}

func saveResult([]HttpInfo) {
	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println(err)
	}

	f, err := os.OpenFile(outputJSONFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	f.WriteString(string(output))

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

	result = append(result, *info)
	fmt.Printf("%-5d %-50s %-60s %s\n", info.StatusCode, info.Url, info.Server, info.Title)
}
