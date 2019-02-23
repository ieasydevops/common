package main

import (
	"sync"
	"math/rand"
	"fmt"
	"time"
)

type Product struct {
	name  int
	value int
}


func producer(wg *sync.WaitGroup, products chan<- Product, name int, stop *bool) {

	for !*stop {
		product := Product{name: name, value: rand.Int()}
		products <- product
		fmt.Printf("producer %v produce a product: %#v\n", name, product)
		time.Sleep(time.Duration(200+rand.Intn(1000)) * time.Millisecond)
	}
	wg.Done()
}


func consumer(wg *sync.WaitGroup, products <-chan Product, name int) {
	for product := range products {
		fmt.Printf("consumer %v consume a product: %#v\n", name, product)
		time.Sleep(time.Duration(200+rand.Intn(1000)) * time.Millisecond)
	}
	wg.Done()
}

func main()  {

	var wgp sync.WaitGroup
	var wgc sync.WaitGroup



	stop := false
	products := make(chan Product, 10)

	// 创建 5 个生产者和 5 个消费者
	for i := 0; i < 5; i++ {
		go producer(&wgp, products, i, &stop)
		go consumer(&wgc, products, i)
		wgp.Add(1)
		wgc.Add(1)
	}

	time.Sleep(time.Duration(10) * time.Second)
	stop = true     // 设置生产者终止信号
	wgp.Wait()      // 等待生产者退出
	close(products) // 关闭通道
	wgc.Wait()      // 等待消费者退出

}