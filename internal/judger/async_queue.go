// Package judger contains pluggable judge implementations.
// judger 包存放可插拔的判题实现。
package judger

import (
	"context"
	"log"
	"sync/atomic"

	"online-judge-backend/internal/services"
)

// AsyncQueue is an in-memory worker pool used by the async judging pipeline.
// AsyncQueue 是异步判题链路使用的内存队列与工作池实现。
type AsyncQueue struct {
	processor services.SubmissionProcessor
	jobs      chan services.SubmissionTask
	workers   int
	busy      int64
}

// NewAsyncQueue creates an in-memory queue with a fixed worker pool.
// NewAsyncQueue 创建带固定工作线程数的内存队列。
func NewAsyncQueue(processor services.SubmissionProcessor, workers, queueCapacity int) *AsyncQueue {
	if workers <= 0 {
		workers = 1
	}
	if queueCapacity <= 0 {
		queueCapacity = workers * 16
	}

	queue := &AsyncQueue{
		processor: processor,
		jobs:      make(chan services.SubmissionTask, queueCapacity),
		workers:   workers,
	}

	for i := 0; i < workers; i++ {
		go queue.runWorker()
	}
	return queue
}

// Enqueue pushes a submission task into the queue.
// Enqueue 把提交任务压入异步队列。
func (q *AsyncQueue) Enqueue(ctx context.Context, task services.SubmissionTask) error {
	select {
	case q.jobs <- task:
		return nil
	case <-ctx.Done():
		return services.ErrJudgeQueueUnavailable
	}
}

// Stats exposes queue depth and worker utilization.
// Stats 暴露队列深度和工作线程利用率。
func (q *AsyncQueue) Stats() services.JudgeQueueStats {
	return services.JudgeQueueStats{
		WorkerCount:   q.workers,
		BusyWorkers:   int(atomic.LoadInt64(&q.busy)),
		QueueDepth:    len(q.jobs),
		QueueCapacity: cap(q.jobs),
	}
}

func (q *AsyncQueue) runWorker() {
	for task := range q.jobs {
		atomic.AddInt64(&q.busy, 1)
		if err := q.processor.Process(context.Background(), task); err != nil {
			log.Printf("judge worker process submission %d: %v", task.SubmissionID, err)
		}
		atomic.AddInt64(&q.busy, -1)
	}
}
