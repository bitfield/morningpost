package main

import (
	"fmt"
	"sync"

	"github.com/thiagonache/morningpost"
	"github.com/thiagonache/morningpost/sources"
)

func main() {
	tg := morningpost.NewTGClient()
	hn := morningpost.NewHNClient()
	tc := morningpost.NewTechCrunchClient()
	bit := sources.NewBITClient()
	var wg sync.WaitGroup
	for _, source := range []morningpost.Source{tg, hn, tc, bit} {
		wg.Add(1)
		src := source
		go func(src morningpost.Source) {
			defer wg.Done()
			news, err := src.GetNews()
			if err != nil {
				fmt.Printf("Cannot get news: %+v\n", err)
				return
			}
			for _, n := range news {
				fmt.Println(n)
			}
		}(src)
	}
	wg.Wait()
}
