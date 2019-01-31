package common

import (
	"flag"
	"fmt"
	"strings"
)

type Msg struct {
	Info  map[string]string
	Count int
}

func (m Msg) Show() {
	for k, v := range m.Info {
		if v != "" {
			fmt.Printf("%s: %s\n", k, v)
		}
	}
	fmt.Printf("Number of scans: %d\n\n", m.Count)
}

func PrintInfo(scanList []string) {
	var n int
	var lines []string
	flag.Visit(func(flag *flag.Flag) {
		if flag.Value.String() != "" {
			line := fmt.Sprintf("%s: %s", flag.Name, flag.Value)
			lines = append(lines, line)

		}
	})
	lines = append(lines, fmt.Sprintf("Number of scans: %d", len(scanList)))

	for _, line := range lines {
		if len(line) >= n {
			n = len(line)
		}
	}

	fmt.Println(strings.Repeat("=", n))
	for _, line := range lines {
		// f := fmt.Sprintf("%%-%ds|\n", n)
		fmt.Println(line)
	}
	fmt.Println(strings.Repeat("=", n))
}
