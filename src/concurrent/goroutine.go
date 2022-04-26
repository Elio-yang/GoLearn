package concurrent

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	//message := make(chan int, 1)
	//ret := make(chan int, 1)
	//message <- 1
	//go spinner(100*time.Millisecond, message, ret)
	//const n = 5
	//fibN := fib(n)
	//fmt.Printf("\rFib(%d)=%d\n", n, fibN)
	//// will block till get
	//fmt.Printf("\r received %d end main \n", <-ret)
	//============================================================
	buffer := make(chan string, 2)
	buffer <- "buffer1"
	buffer <- "buffer2"
	fmt.Printf(<-buffer + "\n")
	fmt.Printf(<-buffer + "\n")
	close(buffer)
	//============================================================
	pings := make(chan string, 1)
	pongs := make(chan string, 1)
	ping(pings, "ping")
	pong(pings, pongs)
	fmt.Printf(<-pongs + "\n")
	close(pings)
	close(pongs)
	//============================================================
	c1 := make(chan string)
	c2 := make(chan string)

	go func() {
		time.Sleep(time.Second)
		c1 <- "from one"
	}()

	go func() {
		time.Sleep(time.Second * 2)
		c2 <- "from two"
	}()

	for i := 0; i < 2; i++ {
		select {
		case msg1 := <-c1:
			fmt.Printf("receive " + msg1 + "\n")
		case msg2 := <-c2:
			fmt.Printf("receive " + msg2 + "\n")
		}
	}
	//============================================================
	go func() {
		time.Sleep(time.Second * 2)
		c1 <- "result 1"
	}()
	for {
		select {
		case res := <-c1:
			fmt.Println(res)
			goto outside
		case <-time.After(time.Second):
			fmt.Println("timeout 1")
		}
	}
outside:
	close(c1)
	close(c2)
	//============================================================
	timer1 := time.NewTimer(time.Second * 2)
	// block until clock arrive
	<-timer1.C
	fmt.Println("Timer 1 arrived")

	timer2 := time.NewTimer(time.Second)
	go func() {
		<-timer2.C
		fmt.Println("Timer 2 arrived")
	}()
	// timer can be cancelled
	stop := timer2.Stop()
	if stop {
		fmt.Println("Timer 2 stopped")
	}
	//============================================================
	// ticker : T=500 ms
	ticker := time.NewTicker(time.Millisecond * 500)
	go func() {
		for t := range ticker.C {
			fmt.Println("tick at ", t)
		}
	}()
	time.Sleep(time.Millisecond * 1600)
	ticker.Stop()
	fmt.Println("timer stopped")
	//============================================================
	jobs := make(chan int, 100)
	results := make(chan int, 100)
	// start 3 worker all blocked
	for w := 1; w < 3; w++ {
		go worker(w, jobs, results)
	}
	// delivery jobs
	for i := 1; i <= 9; i++ {
		jobs <- i
	}
	close(jobs)
	// receive results
	for a := 1; a <= 9; a++ {
		<-results
	}
	//============================================================
	// wait group used for waiting go routines
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func(id int, wg *sync.WaitGroup) {
			fmt.Printf("Worker %d starting\n", id)
			// 睡眠一秒钟，以此来模拟耗时的任务。
			time.Sleep(time.Second)
			fmt.Printf("Worker %d done\n", id)
			// 通知 WaitGroup ，当前协程的工作已经完成。
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	//============================================================
	requests := make(chan int, 5)
	for i := 0; i < 5; i++ {
		requests <- i
	}
	close(requests)

	// 其实就是一个时间点 从当前开始 200ms为周期递增
	limiter := time.Tick(time.Millisecond * 200)
	for req := range requests {
		fmt.Println("request ", req, "<>", <-limiter, "<>", time.Now())
	}

	fmt.Println("========================")

	// 存的是时间点
	bursty := make(chan time.Time, 3)
	for i := 0; i < 3; i++ {
		bursty <- time.Now()
	}
	go func() {
		for t := range time.Tick(time.Millisecond * 200) {
			bursty <- t
		}
	}()
	bursty_requests := make(chan int, 5)
	for i := 0; i < 5; i++ {
		bursty_requests <- i
	}
	close(bursty_requests)
	time.Sleep(time.Second)
	for req := range bursty_requests {
		fmt.Println("request ", req, "<>", <-bursty, "<>", time.Now())
	}
	//============================================================
	var op uint64 = 0
	for i := 0; i < 50; i++ {
		go func() {
			for {
				atomic.AddUint64(&op, 1)
				// 允许其他goroutine
				runtime.Gosched()
			}
		}()
	}
	time.Sleep(time.Second)
	ops := atomic.LoadUint64(&op)
	fmt.Println("ops = ", ops)
	fmt.Println("========================")
	//============================================================
	var state = make(map[int]int)
	var mutex = &sync.Mutex{}
	var opp int64 = 0
	for i := 0; i < 100; i++ {
		go func() {
			total := 0
			for {
				key := rand.Intn(5)
				mutex.Lock()
				total += state[key]
				mutex.Unlock()
				atomic.AddInt64(&opp, 1)

				runtime.Gosched()
			}
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for {
				key := rand.Intn(5)
				val := rand.Intn(100)
				mutex.Lock()
				state[key] = val
				mutex.Unlock()
				atomic.AddInt64(&opp, 1)
				runtime.Gosched()
			}
		}()
	}

	time.Sleep(time.Second)
	opps := atomic.LoadInt64(&opp)
	fmt.Println("opps = ", opps)

	mutex.Lock()
	fmt.Println("state :", state)
	mutex.Unlock()
}

func worker(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Println("worker ", id, "process job ", j)
		time.Sleep(time.Second)
		results <- j * 2
	}
}

func spinner(delay time.Duration, msg chan int, ret chan int) {
	// endless loop
	for {
		for _, r := range `-\|/` {
			fmt.Printf("\r%c", r)
			time.Sleep(delay)
		}
		select {
		case x := <-msg:
			fmt.Printf("\r received %d end go routine\n", x)
			ret <- 2
			break
		default:
			continue
		}
	}
}

func fib(x int) int {
	if x < 2 {
		return x
	} else {
		return fib(x-1) + fib(x-2)
	}
}

// ping is a channel only send
func ping(ping chan<- string, msg string) {
	ping <- msg
}

// ping is used for only receive
func pong(ping <-chan string, pongs chan<- string) {
	msg := <-ping
	pongs <- msg
}
