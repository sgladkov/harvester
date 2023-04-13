package main

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

func main() {
	endpoint := flag.String("a", "localhost:8080", "endpoint to start server (localhost:8080 by default)")
	pollInterval := flag.Int("p", 2, "poll interval")
	reportInterval := flag.Int("r", 10, "report interval")
	flag.Parse()
	if !strings.HasPrefix(*endpoint, "http://") && !strings.HasPrefix(*endpoint, "https://") {
		*endpoint = "http://" + *endpoint
	}

	m := NewMetrics(*endpoint)
	pollTicker := time.NewTicker(time.Duration(*pollInterval) * time.Second)
	go func() {
		for range pollTicker.C {
			err := m.Poll()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println("Metrics are read")
		}
	}()
	reportTicker := time.NewTicker(time.Duration(*reportInterval) * time.Second)
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
}
