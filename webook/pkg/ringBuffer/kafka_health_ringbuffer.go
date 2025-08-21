package ringBuffer

import (
	"math/bits"
	"sync"
)

// Bit1024HealthBuffer 1024比特位的环形缓冲区，记录成功/失败状态
type Bit1024HealthBuffer struct {
	buf   [128]byte // 128字节 = 1024比特
	head  int       // 读指针（0-1023）
	tail  int       // 写指针（0-1023）
	count int       // 当前样本数（0-1024）
	mutex sync.Mutex
}

// NewBit1024HealthBuffer 初始化缓冲区
func NewBit1024HealthBuffer() *Bit1024HealthBuffer {
	return &Bit1024HealthBuffer{}
}

// Push 写入一个状态（success=true记为1，false记为0）
func (b *Bit1024HealthBuffer) Push(success bool) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// 若缓冲区已满，移动读指针（覆盖最旧数据）
	if b.count == 1024 {
		b.head = (b.head + 1) % 1024
	} else {
		b.count++ // 只有未满时才增加计数
	}

	// 计算字节索引和比特位置
	byteIdx := b.tail / 8
	bitPos := b.tail % 8

	// 修复：确保位操作正确
	if success {
		b.buf[byteIdx] |= 1 << bitPos // 置位
	} else {
		b.buf[byteIdx] &^= 1 << bitPos // 清位
	}

	// 移动写指针
	b.tail = (b.tail + 1) % 1024
}

// SuccessRate 计算成功率
func (b *Bit1024HealthBuffer) SuccessRate() float64 {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.count == 0 {
		return 1.0 // 无数据默认100%
	}

	success := 0

	// 修复：正确处理区间计算
	if b.head < b.tail {
		// 连续区间：[head, tail)
		success = b.countBitsInRange(b.head, b.tail-1)
	} else if b.head > b.tail {
		// 环形区间：[head, 1023] + [0, tail-1]
		success += b.countBitsInRange(b.head, 1023)
		success += b.countBitsInRange(0, b.tail-1)
	} else {
		// 读写指针重合且有数据（缓冲区满）
		success = b.countBitsInRange(0, 1023)
	}

	return float64(success) / float64(b.count)
}

// countBitsInRange 统计[start, end]区间内1的个数
func (b *Bit1024HealthBuffer) countBitsInRange(start, end int) int {
	if start > end {
		return 0
	}

	total := 0
	startByte := start / 8
	endByte := end / 8

	// 修复：处理完整字节的统计逻辑
	if startByte == endByte {
		// 同一字节内的部分比特
		startBit := start % 8
		endBit := end % 8
		mask := (byte(0xFF) << startBit) & (byte(0xFF) >> (7 - endBit))
		total += bits.OnesCount8(b.buf[startByte] & mask)
	} else {
		// 起始字节的部分比特
		startBit := start % 8
		if startBit < 8 {
			mask := byte(0xFF) << startBit
			total += bits.OnesCount8(b.buf[startByte] & mask)
		}

		// 中间完整字节
		for i := startByte + 1; i < endByte; i++ {
			total += bits.OnesCount8(b.buf[i])
		}

		// 结束字节的部分比特
		endBit := end % 8
		mask := byte(0xFF) >> (7 - endBit)
		total += bits.OnesCount8(b.buf[endByte] & mask)
	}

	return total
}

// IsHealthy 判断是否健康
func (b *Bit1024HealthBuffer) IsHealthy(minSuccessRate float64, minSamples int) bool {
	b.mutex.Lock()
	count := b.count
	b.mutex.Unlock()

	if count < minSamples {
		return true
	}

	return b.SuccessRate() >= minSuccessRate
}
