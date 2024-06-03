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

func RepeatFn(done <-chan struct{}, fn func() int) <-chan int {
	repeatStream := make(chan int)
	go func() {
		defer close(repeatStream)
		for {
			select {
			case <-done:
				return
			case repeatStream <- fn():
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

func OrDone(done <-chan struct{}, inputStream <-chan int) <-chan int {
	outputStream := make(chan int)
	go func() {
		defer close(outputStream)
		for {
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
	}()
	return outputStream
}

func Tee(done <-chan struct{}, inputStream <-chan int) (<-chan int, <-chan int) {
	outputStream1 := make(chan int)
	outputStream2 := make(chan int)
	go func() {
		defer close(outputStream1)
		defer close(outputStream2)
		for v := range OrDone(done, inputStream) {
			out1, out2 := outputStream1, outputStream2
			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				case out1 <- v:
					out1 = nil
				case out2 <- v:
					out2 = nil
				}
			}
		}
	}()
	return outputStream1, outputStream2
}

func Bridge(done <-chan struct{}, inputStreams <-chan (<-chan int)) <-chan int {
	outputStream := make(chan int)
	go func() {
		defer close(outputStream)
		for {
			var stream <-chan int
			select {
			case <-done:
				return
			case inputStream, ok := <-inputStreams:
				if !ok {
					return
				}
				stream = inputStream
			}
			for v := range OrDone(done, stream) {
				select {
				case <-done:
					return
				case outputStream <- v:
				}
			}
		}
	}()
	return outputStream
}
