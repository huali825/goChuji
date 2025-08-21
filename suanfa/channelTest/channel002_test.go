package channelTest

import (
	"fmt"
	"testing"
	"time"
)

func doTaskWithTimeout(task func(), timeout time.Duration) error {
	done := make(chan struct{})

	go func() {
		task()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return fmt.Errorf("超时错误")
	}
}

func TestChannel002(t *testing.T) {
	err := doTaskWithTimeout(
		func() {
			time.Sleep(1 * time.Second)
			fmt.Println("任务完成")
		}, 2*time.Second,
	)

	fmt.Println(err)
}
