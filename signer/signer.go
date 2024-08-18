package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

var (
	SingleHash = func(in, out chan interface{}) {
		wg := &sync.WaitGroup{}
		mt := &sync.Mutex{}
		for data := range in {
			wg.Add(1)
			go func(wg *sync.WaitGroup, mt *sync.Mutex, data int, out chan interface{}) {
				defer wg.Done()
				convertedData := fmt.Sprintf("%d", data)
				l := make(chan string)
				r := make(chan string)

				go func(l chan string, data string) {
					l <- DataSignerCrc32(data)
				}(l, convertedData)

				go func(r chan string, data string) {
					mt.Lock()
					md5 := DataSignerMd5(data)
					mt.Unlock()
					r <- DataSignerCrc32(md5)
				}(r, convertedData)

				rData := <-r
				lData := <-l

				result := lData + "~" + rData
				out <- result
			}(wg, mt, data.(int), out)
		}

		wg.Wait()
	}

	MultiHash = func(in, out chan interface{}) {
		wgOuter := &sync.WaitGroup{}
		for data := range in {
			wgOuter.Add(1)
			go func(wgOuter *sync.WaitGroup, data string, out chan interface{}) {
				defer wgOuter.Done()
				var results = make([]string, 6)
				mt := &sync.Mutex{}
				wgInner := &sync.WaitGroup{}

				for th := 0; th < 6; th++ {
					wgInner.Add(1)
					go func(wgInner *sync.WaitGroup, mt *sync.Mutex, results []string, th int) {
						defer wgInner.Done()
						r := fmt.Sprintf("%d%s", th, data)
						hash := DataSignerCrc32(r)
						mt.Lock()
						results[th] = hash
						mt.Unlock()
					}(wgInner, mt, results, th)
				}

				wgInner.Wait()
				result := strings.Join(results, "")
				out <- result
			}(wgOuter, data.(string), out)
		}
		wgOuter.Wait()
	}

	CombineResults = func(in, out chan interface{}) {
		var results []string

		for r := range in {
			results = append(results, r.(string))
		}

		sort.Strings(results)
		out <- strings.Join(results, "_")
	}
)

func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	chans := make([]chan interface{}, len(jobs)+1)

	for i := 1; i < len(chans); i++ {
		chans[i] = make(chan interface{})
	}

	for i, j := range jobs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, j job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			j(in, out)
		}(wg, j, chans[i], chans[i+1])
	}

	wg.Wait()
}
