package main

import (
	"fmt"
	"github.com/ekoops/goncurrency/pkg/pipeline"
	"runtime"
	"time"
)

func unlockingTest() {
	done := make(chan struct{})
	go func() {
		time.Sleep(1 * time.Second)
		close(done)
	}()
	for range pipeline.OrDone(done, nil) {
		fmt.Println("Unreachable code")
	}
}

func main() {
	done := make(chan struct{})
	defer close(done)

	inputStream := pipeline.Repeat(done, 1, 2)

	fmt.Println("Starting base pipeline test...")
	start := time.Now()
	for v := range pipeline.Take(done, pipeline.Heavy(done, inputStream, time.Second), 6) {
		fmt.Println(v)
	}
	fmt.Println("Time elapsed (base pipeline):", time.Since(start))

	numStreams := runtime.NumCPU()
	fmt.Printf("Generating %d streams...\n", numStreams)
	inputStreams := make([]<-chan int, numStreams)
	for i := range inputStreams {
		inputStreams[i] = pipeline.Heavy(done, inputStream, time.Second)
	}

	fmt.Println("Starting fan-out/fan-in pattern test...")
	start = time.Now()
	for v := range pipeline.Take(done, pipeline.FanInUnordered(done, inputStreams...), 6) {
		fmt.Println(v)
	}
	fmt.Println("Time elapsed (fan-out/fan-in pattern):", time.Since(start))

	fmt.Println("Starting unlocking test...")
	unlockingTest()
	fmt.Println("Successfully unlocked")
}
