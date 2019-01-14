package common

import (
	"flag"
)

type Options struct {
	Host        *string
	Port        *string
	File        *string
	Timeout     *int
	Threads     *int
	OutputFile  *string
	ShowVerbose *bool
}

var PublicOptions = Options{
	Host:       flag.String("host", "", "Host or Host Range. 127.0.0.1 | 192.168.1.1/24 | 192.168.1.1-5"),
	Port:       flag.String("p", "", "Port or Port Range. 80. 1-65535 | 21,22,25 | 8080"),
	File:       flag.String("f", "", "Load File Path"),
	Timeout:    flag.Int("timeout", 1, "Connection timeout"),
	Threads:    flag.Int("t", 200, "Number of concurrent threads"),
	OutputFile: flag.String("o", "", "Result output file path"),
}
