package util

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// JobResp 每个shopapi请求返回结果
type JobResp struct {
	resp *http.Response
	err  error
}

// Job 每个shopapi请求封装
type Job struct {
	req        string
	resultChan chan *JobResp
	timeout    time.Time // 超时时间，如果请求没有在这个时间点内返回，则按超时算
	next       *Job
}

// NewJob 创建job对象
func NewJob(req string, timeout time.Time) (job *Job) {
	job = &Job{
		req:        req,
		resultChan: make(chan *JobResp, 1),
		timeout:    timeout,
	}
	return job
}

// JobQueue 任务队列
type JobQueue struct {
	maxSize int // 0 表示无上限
	cond    *sync.Cond
	size    int64 // 队列的长度
	stop    int64 // 标记是否停止
	head    *Job
	tail    *Job
}

// NewJobQueue create jobqueue
func NewJobQueue(maxSize int) (jq *JobQueue) {
	mu := &sync.Mutex{}
	jq = &JobQueue{
		maxSize: maxSize,
		cond:    sync.NewCond(mu),
	}
	return
}

// Stop 停止队列
func (jq *JobQueue) Stop() {
	atomic.StoreInt64(&jq.stop, 1)
}

// IsStop 判断是否停止
func (jq *JobQueue) IsStop() bool {
	return atomic.LoadInt64(&jq.stop) > 0
}

// Size 队列长度
func (jq *JobQueue) Size() (size int64) {
	jq.cond.L.Lock()
	size = jq.size
	jq.cond.L.Unlock()
	return
}

// Push 添加到队尾部，返回false表示列表已满
func (jq *JobQueue) Push(job *Job) (succ bool) {

	jq.cond.L.Lock()

	if jq.IsStop() {
		jq.cond.L.Unlock()
		return false
	}
	// 队列过长
	if jq.maxSize != 0 && jq.size >= int64(jq.maxSize) {
		jq.cond.L.Unlock()
		return false
	}

	jq.size++
	if jq.head == nil && jq.tail == nil {
		jq.head = job
		jq.tail = job
		job.next = nil
	} else {
		job.next = nil
		jq.tail.next = job
		jq.tail = job
	}
	jq.cond.L.Unlock()
	// 通知pop等待的去取job
	jq.cond.Signal()
	return true
}

// Pop 从队列里取内容，队列为空则等待
func (jq *JobQueue) Pop() (job *Job) {

	jq.cond.L.Lock()

	// 一直等待有内容为止
	for jq.size == 0 && !jq.IsStop() {
		jq.cond.Wait()
	}

	if jq.size == 0 && jq.IsStop() {
		jq.cond.L.Unlock()
		jq.cond.Broadcast() // 通知所有人队列结束
		return nil
	}
	job = jq.head
	jq.head = jq.head.next
	jq.size--
	if jq.head == nil {
		jq.tail = nil
	}
	jq.cond.L.Unlock()
	return
}
