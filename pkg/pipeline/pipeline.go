package pipeline

import (
	"sync"
	"time"
)

func Repeat(done <-chan struct{}, integers ...int) <-chan int {
	repeatStream := make(chan int)
	go func() {
		defer close(repeatStream)
		for {
			for _, i := range integers {
				select {
				case <-done:
					return
				case repeatStream <- i:
				}
			}
		}
	}()
	return repeatStream
}

func Take(done <-chan struct{}, inputStream <-chan int, n int) <-chan int {
	takeStream := make(chan int)
	go func() {
		defer close(takeStream)
		for i := 0; i < n; i++ {
			select {
			case <-done:
				return
			case v, ok := <-inputStream:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				case takeStream <- v:
				}
			}
		}
	}()
	return takeStream
}

func Heavy(done <-chan struct{}, inputStream <-chan int, d time.Duration) <-chan int {
	heavyStream := make(chan int)
	go func() {
		defer close(heavyStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-inputStream:
				if !ok {
					return
				}
				time.Sleep(d)
				select {
				case <-done:
					return
				case heavyStream <- v:
				}
			}
		}
	}()
	return heavyStream
}

func multiplex(done <-chan struct{}, wg *sync.WaitGroup, inputStream <-chan int, outputStream chan<- int) {
	defer wg.Done()
	select {
	case <-done:
		return
	case v, ok := <-inputStream:
		if !ok {
			return
		}
		select {
		case <-done:
			return
		case outputStream <- v:
		}
	}
}

func FanInUnordered(done <-chan struct{}, inputStreams ...<-chan int) <-chan int {
	fanInStream := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(inputStreams))
	for _, inputStream := range inputStreams {
		go multiplex(done, &wg, inputStream, fanInStream)
	}
	go func() {
		defer close(fanInStream)
		wg.Wait()
	}()
	return fanInStream
}
