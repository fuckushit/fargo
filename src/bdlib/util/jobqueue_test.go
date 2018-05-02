package util

import (
	"fmt"
	"runtime"
	// "sync/atomic"
	"testing"
	"time"
)

var (
	sentinel = fmt.Errorf("sentinel")
)

func TestJobQueue(t *testing.T) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	jq := NewJobQueue(0)

	// 起一堆goroutine去做消费者
	for i := 0; i < 10; i++ {
		go consumer(jq, i)
	}

	var array []*Job
	for i := 0; i < 100; i++ {
		job := NewJob("model", time.Now().Add(time.Second))
		jq.Push(job)
		array = append(array, job)
	}
	fmt.Println("jobqueue stop")
	jq.Stop()
	time.Sleep(2 * time.Second)
	for _, job := range array {
		resp := <-job.resultChan
		if resp.err != sentinel {
			t.Fatal("result is illegal")
		}
	}
}

func consumer(jq *JobQueue, index int) {
	for {
		job := jq.Pop()
		if job == nil {
			fmt.Printf("consumer %d quit\n", index)
			return
		}
		fmt.Printf("<<<<<<<<<< consumer %d handle job %s\n", index, job.req)
		resp := &JobResp{
			err: sentinel,
		}
		time.Sleep(10 * time.Millisecond)
		job.resultChan <- resp
	}
}

// func TestJobQueuePerformance(t *testing.T) {
// 	runtime.GOMAXPROCS(runtime.NumCPU())
// 	var counter int64
// 	jq := NewJobQueue(0)
// 	for i := 0; i < 100; i++ {
// 		go func() {
// 			for {
// 				job := NewJob("user", time.Now().Add(10*time.Second))
// 				fmt.Println(">>>>>>>>>> job push:%v, i:%v", job, i)
// 				jq.Push(job)
// 			}
// 		}()
// 	}
// 	for i := 0; i < 100; i++ {
// 		go func() {
// 			for {
// 				job := jq.Pop()
// 				fmt.Println("<<<<<<<<<< job pop:%v, i:%v", job, i)
// 				atomic.AddInt64(&counter, 1)
// 			}
// 		}()
// 	}

// 	ticker := time.NewTicker(1 * time.Second)
// 	for _ = range ticker.C {
// 		num := atomic.SwapInt64(&counter, 0)
// 		fmt.Printf("qps %d queue_len %d\n", num, jq.Size())
// 	}
// }
