package suanfa

import (
	"fmt"
	"testing"
)

func TestSuanfa004(t *testing.T) {
	temp := ("leetcode")
	fmt.Println(temp)
}

// 编程题 : 寻找字符串中的第一个唯一字符
// 题目描述：
// 给定一个字符串 s，找到它的第一个只出现一次的字符，并返回它的索引。如果不存在，则返回 -1。
// 示例：
// - s = "leetcode" -> 返回 0 (因为 l 是第一个只出现一次的字符)
// - s = "loveleetcode" -> 返回 2 (因为 v 是第一个只出现一次的字符)
// - s = "aabb" -> 返回 -1
func findOnceFirstChar(s string) int {
	count := make(map[rune]int)
	for _, char := range s {
		count[char]++
	}
	for k, char := range s {
		if count[char] == 1 {
			return k
		}
	}
	return -1
}
