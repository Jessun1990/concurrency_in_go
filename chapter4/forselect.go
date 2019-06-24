package chapter4

import (
	"fmt"
	"math/rand"
	"time"
)

/* P103
for { // 无限循环或者使用 range 循环
	select {
		// 使用 channel 进行作业
	}
}
*/

/*
for _, s := range []string{"a", "b", "c"}{
	select {
		case <- done :
			return
		case stringStream <- s:
	}
}
*/

/*
循环等待停止
第一种
for {
	select {
		case <- done:
			return
		default:
	}
	// 进行抢占式任务
}
第二种
for {
	select {
		case <- done:
			return
		default:
			// 进行抢占式任务
	}
}
*/

// goroutineExample: 简单的 goroutine 泄漏举例
func goroutineExample() {
	doWork := func(strings <-chan string) <-chan interface{} {
		completed := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(completed)
			for s := range strings {
				fmt.Println(s)
			}
		}()
		return completed
	}
	doWork(nil)
	fmt.Println("Done.")
}

// goroutineExample2 使用 done 信号来通知 goroutine 退出，P106
// channel 上接收 goroutine 。
func goroutineExample2() {
	doWork := func(
		done <-chan interface{}, strings <-chan interface{},
	) <-chan interface{} {
		terminated := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited.")
			defer close(terminated)
			for {
				select {
				case s := <-strings:
					fmt.Println(s)
				case <-done:
					return
				}
			}
		}()
		return terminated
	}

	done := make(chan interface{})
	terminated := doWork(done, nil)

	go func() {
		fmt.Println("Canceling doWork goroutine ...")
		time.Sleep(10 * time.Second)
		close(done)
	}()
	<-terminated
	fmt.Println("Done.")
}

// goroutine 阻塞了向 channel 进行写入的请求， P108
func goroutineExample3() {
	newRandStream := func(done <-chan interface{}) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited")
			defer close(randStream)

			for {
				select {
				case randStream <- rand.Int():
				case <-done:
					return
				}
			}
		}()
		return randStream
	}

	done := make(chan interface{})
	randStream := newRandStream(done)
	fmt.Println("3 random ints:")
	for i := 1; i <= 3; i++ {
		fmt.Printf("%+v: %+v\n", i, <-randStream)
	}
	close(done)
	time.Sleep(time.Second)
}

// P109 通过递归和 goroutine 创建一个复合的 done channel，将一个或者多个完成的 channel 合并到一个完成的 channel 中。
// 将任何数量的 channel 组合到单个 channel 中，只要任何组件 channel 关闭或写入，该 channel 就会关闭，依赖于其关闭的 channel 也将关闭。
func goroutineExample4() {
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil // 递归函数的终止标准
		case 1:
			return channels[0]
		}

		orDone := make(chan interface{})
		go func() {
			defer close(orDone)

			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}
	// ================== 以上是定义

	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}
	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v", time.Since(start))
}
