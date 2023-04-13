package main

import (
	"fmt"
	"time"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	m := NewMetrics("http://localhost:8080")
	pollInterval := 2
	pollTicker := time.NewTicker(time.Duration(pollInterval) * time.Second)
	go func() {
		for range pollTicker.C {
			err := m.Poll()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Metrics are read")
		}
	}()
	reportInterval := 10
	reportTicker := time.NewTicker(time.Duration(reportInterval) * time.Second)
	go func() {
		for range reportTicker.C {
			err := m.Report()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Metrics are reported")
		}
	}()
	//r := bufio.NewReader(os.Stdin)
	//fmt.Println("Press Enter to exit")
	//r.ReadLine()
	for {
	}
	pollTicker.Stop()
	reportTicker.Stop()
	return nil
}
