package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/hansumane/go-metar/pkg/fetcher"
	"github.com/hansumane/go-metar/pkg/parser"
)

func main() {
	airports := parser.StringsUpper(os.Args[1:])
	processedMetars := make(map[string]bool)

	for {
		metars, err := fetcher.FetchMetars(airports)
		if err != nil {
			log.Println(err)
			time.Sleep(time.Second)
			continue
		}
		if len(metars) != len(airports) {
			log.Println("not enough metars fetched")
			time.Sleep(time.Second)
			continue
		}

		var newMetars []parser.Metar
		for _, raw := range metars {
			if _, ok := processedMetars[raw]; ok {
				continue
			}
			processedMetars[raw] = true
			newMetars = append(newMetars, parser.NewMetar(raw))
		}

		for _, m := range newMetars {
			fmt.Printf("DEBUG: %s\n", m.Raw)
		}
		if len(newMetars) != 0 {
			fmt.Printf("\n")
		}

		for _, m := range newMetars {
			if err := m.Parse(); err != nil {
				log.Println(err)
			} else {
				fmt.Printf("%s\n", m)
			}
		}
		if len(newMetars) != 0 {
			fmt.Println()
		}

		time.Sleep(2 * time.Minute)
	}
}
