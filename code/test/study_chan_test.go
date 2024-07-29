package test

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"testing"
)

func TestMain(t *testing.T) {
	fmt.Println("test")
	fmt.Println(string(os.PathSeparator))
	fmt.Println(32 << 20)
}

// 一直等待直到WaitGroup等于0
func TestSyncWait(t *testing.T) {
	var vg sync.WaitGroup
	vg.Add(1)
	go func() {
		defer vg.Done()
		fmt.Println("hello")
	}()
	vg.Wait()
}

// 一个已经锁了的锁，再锁一次会一直阻塞，这个不建议使用
func TestSyncMutex(t *testing.T) {
	var m sync.Mutex
	m.Lock()
	// m.Lock()
}

// 空Select，没有case会一直阻塞
func TestSyncSelect(t *testing.T) {
	select {}
}

// 死循环，不建议使用，因为100%会占用一个cpu
func TestSyncFor(t *testing.T) {
	for {
	}
}

// 系统信号量  在go内也是一个channel，在收到特定的消息前一直阻塞
func TestSyncSemaphore(t *testing.T) {
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	<-sig
}

// 空channel
func TestSyncChannel(t *testing.T) {
	c := make(chan struct{})
	<-c
}

// nil channel
func TestSyncChannel2(t *testing.T) {
	var c chan struct{} // nil channel
	<-c
}
