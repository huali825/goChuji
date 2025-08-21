package channelTest

import (
	"fmt"
	"sync"
	"testing"
)

// Singleton 是一个单例结构体，包含一个 name 字段
type Singleton struct {
	name string // name 是 Singleton 结构体的一个字段，用于存储实例名称
}

// 使用 var 声明全局变量，确保单例实例和 once 控制器的唯一性
var (
	singleton *Singleton // singleton 用于存储单例实例的指针
	once      sync.Once  // once 用于确保初始化过程只执行一次
)

// GetInstance 是获取单例实例的公共方法
// 使用 sync.Once 确保单例实例只被创建一次
// 返回单例实例的指针
func GetInstance() *Singleton {
	once.Do(func() {
		// 在这里执行单例的初始化工作
		singleton = &Singleton{name: "singleton"} // 创建单例实例并设置初始值
		fmt.Println("单例实例被创建")
	})
	return singleton // 返回单例实例
}
func (s *Singleton) PrintName() {
	fmt.Println(s.name)
}

func TestChannel001(t *testing.T) {
	var wg sync.WaitGroup

	// 启动10个goroutine同时获取单例
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 获取单例实例
			singleton := GetInstance()
			// 打印实例地址，验证是否为同一个实例
			fmt.Printf("获取到的单例地址: %p\n", singleton)
		}()
	}

	wg.Wait()
}
