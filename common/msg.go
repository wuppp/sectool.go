package common

import (
	"fmt"
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
