package suanfa

import (
	"fmt"
	"testing"
)

func TestSuanfa001(t *testing.T) {
	//defer实例
	b()
}

func b() {
	for i := 0; i < 4; i++ {
		defer fmt.Print(i)
	}
}
