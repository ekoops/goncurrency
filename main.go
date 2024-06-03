package main

import (
	"fmt"
	"github.com/ekoops/goncurrency/pkg/pipeline"
	"math/rand"
	"runtime"
	"time"
)

func main() {
	//testBasePipeline()
	//testFanOutFanIn()
	//testUnlocking()
	//testTee()
	testBridge()
}

func testBasePipeline() {
	done := make(chan struct{})
	defer close(done)

	inputStream := pipeline.Repeat(done, 1, 2)

	fmt.Println("Starting base pipeline test...")
	start := time.Now()
	for v := range pipeline.Take(done, pipeline.Heavy(done, inputStream, time.Second), 6) {
		fmt.Println(v)
	}
	fmt.Println("Time elapsed (base pipeline):", time.Since(start))
}

func testFanOutFanIn() {
	done := make(chan struct{})
	defer close(done)

	inputStream := pipeline.Repeat(done, 1, 2)

	numStreams := runtime.NumCPU()
	fmt.Printf("Generating %d streams...\n", numStreams)
	inputStreams := make([]<-chan int, numStreams)
	for i := range inputStreams {
		inputStreams[i] = pipeline.Heavy(done, inputStream, time.Second)
	}

	fmt.Println("Starting fan-out/fan-in pattern test...")
	start := time.Now()
	for v := range pipeline.Take(done, pipeline.FanInUnordered(done, inputStreams...), 6) {
		fmt.Println(v)
	}
	fmt.Println("Time elapsed (fan-out/fan-in pattern):", time.Since(start))
}

func testUnlocking() {
	done := make(chan struct{})
	fmt.Println("Starting unlocking test...")
	go func() {
		time.Sleep(1 * time.Second)
		close(done)
	}()
	for range pipeline.OrDone(done, nil) {
		fmt.Println("Unreachable code")
	}
	fmt.Println("Successfully unlocked")
}

func testTee() {
	done := make(chan struct{})
	defer close(done)

	fmt.Println("Starting tee test...")
	randGen := func() int { return rand.Intn(1000) }
	out1, out2 := pipeline.Tee(done, pipeline.Take(done, pipeline.RepeatFn(done, randGen), 10))
	for v := range out1 {
		fmt.Printf("out1: %d, out2: %d\n", v, <-out2)
	}
	fmt.Println("Tee test finished")
}

func testBridge() {
	genVals := func(done <-chan struct{}) <-chan (<-chan int) {
		valStreams := make(chan (<-chan int))
		go func() {
			defer close(valStreams)
			for i := 0; i < 10; i++ {
				stream := make(chan int, 1)
				stream <- i
				close(stream)
				valStreams <- stream
			}
		}()
		return valStreams
	}

	done := make(chan struct{})
	defer close(done)

	fmt.Println("Starting bridge test...")
	for v := range pipeline.Bridge(done, genVals(done)) {
		fmt.Println(v)
	}
	fmt.Println("Bridge test finished")
}
