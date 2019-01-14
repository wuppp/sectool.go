package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"golang.org/x/net/html/charset"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

func Fetch(url string) ([]byte, error) {
	resp, err := http.Get(url)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	// reader := bufio.NewReader(resp.Body)
	e := determineEncoding(resp.Body)

	utf8Reader := transform.NewReader(resp.Body, e.NewDecoder())

	body, err := ioutil.ReadAll(utf8Reader)
	if err != nil {
		panic(err)
	}

	return body, nil
}

func determineEncoding(r io.Reader) encoding.Encoding {
	b, err := bufio.NewReader(r).Peek(1024)
	if err != nil {
		fmt.Println("get code error")
		return unicode.UTF8
	}
	e, _, _ := charset.DetermineEncoding(b, "")
	return e
}

func main() {
	// resp, err := http.Get("http://46.29.162.160/1")

	s, _ := Fetch("http://193.187.119.20:80/123")
	fmt.Print(string(s))
}
