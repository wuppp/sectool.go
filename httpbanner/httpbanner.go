package main

import (
	. "common"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

var (
	wg             sync.WaitGroup
	ch             chan bool
	host           string
	port           string
	timeout        int
	outputJSONFile string
	goroutineNum   int
	result         []HttpInfo
)

var headers = map[string]string{
	"X-Type":     "Scan",
	"User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36",
}

type HttpInfo struct {
	StatusCode int    `json:"status_code"`
	Url        string `json:"url"`
	Title      string `json:"title"`
	Server     string `json:"server"`
}

func main() {

	flag.StringVar(&host, "host", "", "scan hosts")
	flag.IntVar(&timeout, "timeout", 5, "http connect timeout")
	flag.StringVar(&port, "port", "80", "scan port. Default 80")
	flag.IntVar(&goroutineNum, "t", 1000, "scan thread number. Default 1000")
	flag.StringVar(&outputJSONFile, "oJ", "", "save result file")
	flag.Parse()

	//限制goroutine数量
	ch = make(chan bool, goroutineNum)

	if host == "" {
		flag.Usage()
		os.Exit(0)
	}

	ipList, _ := ParseIP(host)
	portList, _ := ParsePort(port)
	urls := []string{}

	for _, host := range ipList {
		for _, port := range portList {
			url := fmt.Sprintf("http://%s:%d/", host, port)
			urls = append(urls, url)
		}
	}

	for _, url := range urls {
		ch <- true
		wg.Add(1)
		go fetch(url)
	}
	wg.Wait()
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

	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	for k, v := range headers {
		req.Header.Add(k, v)
	}
	resp, err := client.Do(req)
	if err != nil {
		//log.Println("http.Get:", err.Error())
		return
	}
	defer resp.Body.Close()

	info := &HttpInfo{}
	info.Url = url
	info.StatusCode = resp.StatusCode

	//获取标题
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		//log.Println("ioutil.ReadAll", err.Error())
		return
	}
	respBody := string(body)
	r := regexp.MustCompile(`(?i)<title>\s?(.*?)\s?</title>`)
	m := r.FindStringSubmatch(respBody)
	if len(m) == 2 {
		info.Title = m[1]
	}

	//获取响应头Server字段
	info.Server = resp.Header.Get("Server")

	result = append(result, *info)
	fmt.Printf("%-5d %-29s %-60s %s\n", info.StatusCode, info.Url, info.Server, info.Title)
}
