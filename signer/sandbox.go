package main

import (
	"fmt"
)

func test() {
	ExecutePipeline([]job{
		job(func(in, out chan interface{}) {
			for _, num := range []int{0, 1, 2, 3, 4, 5, 6} {
				out <- num
			}
		}),

		job(SingleHash),
		//job(MultiHash),
		job(CombineResults),

		job(func(in, out chan interface{}) {
			data := <-in
			s := data.(string)
			fmt.Println("Result: ", s)
		}),
	}...)
}

func main() {
	test()
}
