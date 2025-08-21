package ringBuffer

import (
	"fmt"
	"testing"
	"time"
)

func TestNewBit1024HealthBuffer(t *testing.T) {
	// 初始化1024比特的健康度缓冲区
	healthBuf := NewBit1024HealthBuffer()

	// 模拟Kafka发送结果（例如：95%成功率）
	for i := 0; i < 2000; i++ { // 超过1024，触发环形覆盖
		success := i%20 != 0 // 19/20=95%成功率
		healthBuf.Push(success)
	}

	time.Sleep(time.Second)

	// 判断健康状态（要求至少500个样本，成功率≥90%）
	if healthBuf.IsHealthy(0.9, 500) {
		rate := healthBuf.SuccessRate() * 100
		fmt.Printf("Kafka通道健康，成功率: %.2f%%\n", rate)
	} else {
		rate := healthBuf.SuccessRate() * 100
		fmt.Printf("Kafka通道不健康，成功率: %.2f%%\n", rate)
	}
}

func TestBit1024HealthBuffer01(t *testing.T) {
	buf := NewBit1024HealthBuffer()

	// 测试1：部分填充（100成功+10失败）
	for i := 0; i < 100; i++ {
		buf.Push(true)
	}
	for i := 0; i < 10; i++ {
		buf.Push(false)
	}
	rate := buf.SuccessRate()
	expected1 := 100.0 / 110 // ≈90.91%
	if !floatEqual(rate, expected1, 0.01) {
		t.Errorf("测试1失败：预期≈%.2f%%，实际%.2f%%", expected1*100, rate*100)
	}

	// 测试2：填满缓冲区（此时仍保留10个失败，不会覆盖）
	remaining := 1024 - 110 // 需要再写入914个
	for i := 0; i < remaining; i++ {
		buf.Push(true)
	}
	rate = buf.SuccessRate()
	expected2 := 1014.0 / 1024 // 100+914=1014成功，10失败
	if !floatEqual(rate, expected2, 0.01) {
		t.Errorf("测试2失败：预期≈%.2f%%，实际%.2f%%", expected2*100, rate*100)
	}

	// 测试3：覆盖旧数据（写入10个成功，覆盖最初的10个失败）
	// 此时缓冲区已满，写入会覆盖最旧的数据（前10个是F）
	for i := 0; i < 10; i++ {
		buf.Push(true)
	}
	rate = buf.SuccessRate()
	expected3 := 1024.0 / 1024 // 所有失败被覆盖，100%成功
	if !floatEqual(rate, expected3, 0.01) {
		t.Errorf("测试3失败：预期≈%.2f%%，实际%.2f%%", expected3*100, rate*100)
	}

	// 测试4：继续覆盖（写入100个失败，覆盖最旧的100个成功）
	for i := 0; i < 100; i++ {
		buf.Push(false)
	}
	rate = buf.SuccessRate()
	expected4 := (1024 - 100) / 1024.0 // 1024-100=924成功
	if !floatEqual(rate, expected4, 0.01) {
		t.Errorf("测试4失败：预期≈%.2f%%，实际%.2f%%", expected4*100, rate*100)
	}

	fmt.Printf("所有测试通过，最终成功率: %.2f%%\n", rate*100)
}

// 辅助函数：判断两个浮点数是否在误差范围内相等
func floatEqual(a, b, epsilon float64) bool {
	if a-b < 0 {
		return b-a < epsilon
	}
	return a-b < epsilon
}
