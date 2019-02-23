package main

import (
	"context"
	"time"
	"fmt"
	"errors"
	"log"
)

type RunnerLoop struct {
	ctx     context.Context
	recvCtx context.Context
	cancel  func()
	runner  *Runner
	stopped chan struct{}
}

type Runner struct {
	RecvChannel  <-chan []interface{}
	buffer       []interface{}
	timeout      time.Duration
	lastRecvSize int
}

func (this *Runner) DoRecv(ctx context.Context) error {
	var ok bool
	timer := time.NewTicker(this.timeout)
	select {
	case this.buffer, ok = <-this.RecvChannel:
		if !ok {
			fmt.Println("channel closed")
			return errors.New("channel closed")
		}
	case <-timer.C:
		fmt.Println("receive data timeout")
		break
	}
	return nil

}

type Consumer struct {
	commonChannel <-chan interface{}
}

func (this *Consumer) Start() {

}

func NewRunnerLoop(ctx context.Context, r *Runner) *RunnerLoop {
	rl := &RunnerLoop{
		ctx:    ctx,
		runner: r,
	}
	rl.recvCtx, rl.cancel = context.WithCancel(ctx)
	return rl
}

func (this *RunnerLoop) run(interval, timeout time.Duration, errc chan<- error) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
mainLoop:
	for {
		fmt.Println("do Receiver Channel begin")
		select {
		case <-this.ctx.Done():
			close(this.stopped)
			return
		case <-this.recvCtx.Done():
			break mainLoop
		default:
		}

		var (
			recvCtx, cancel = context.WithTimeout(this.ctx, timeout)
		)

		recvErr := this.runner.DoRecv(recvCtx)
		cancel()
		if recvErr == nil {
			// NOTE: There were issues with misbehaving clients in the past
			// that occasionally returned empty results. We don't want those
			// to falsely reset our buffer size.
			if len(this.runner.buffer) > 0 {
				this.runner.lastRecvSize = len(this.runner.buffer)
			}
		} else {
			fmt.Println("Recv failed", "err", recvErr)
			if errc != nil {
				errc <- recvErr
			}
		}
		fmt.Println(this.runner.buffer)
		select {
		case <-this.ctx.Done():
			close(this.stopped)
			return
		case <-this.recvCtx.Done():
			break mainLoop
		case <-ticker.C:
		}
	}
	close(this.stopped)
}


var i=0

func tickerTest(ctx context.Context) {

	var (
		interval = time.Duration(1) * time.Second
	)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
mainLoop1:
	for {

		select {
		case <-ctx.Done():
			return
		default:
		}

		i++
		fmt.Printf("do somthing %d\n",i)

		select {
		case <-ctx.Done():
			break mainLoop1
		case <-ticker.C:
		}
	}
}

func timeoutHandler(ctx context.Context) {
	go doStuff(ctx,time.Duration(1)*time.Second)
}


//每1秒work一下，同时会判断ctx是否被取消了，如果是就退出
func doStuff(ctx context.Context,duration time.Duration) {
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("done")
			return
		case <-ticker.C:
			log.Printf("work")
		}
	}
}



func main() {
	ctx :=context.Background()
	ctx1,CancelFunc :=context.WithTimeout(ctx,time.Duration(15)*time.Second)
	defer CancelFunc()

	timeoutHandler(ctx1)
	tickerTest(ctx)
	select {
	case <-ctx.Done():
		log.Println("AllDone")
	}
}
