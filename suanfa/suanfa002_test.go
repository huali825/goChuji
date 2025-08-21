package suanfa

import (
	"testing"
)

// Definition for singly-linked list.
type ListNode struct {
	Val  int
	Next *ListNode
}

func TestSuanfa002(t *testing.T) {
	//defer实例
	b()
}

// 3. 无重复字符的最长子串
func lengthOfLongestSubstring(s string) int {
	charMap := make(map[byte]int)
	left := 0
	maxLen := 0
	for i := 0; i < len(s); i++ {
		currChar := s[i]
		if idx, ok := charMap[currChar]; ok && idx >= left {
			left = idx + 1
		}
		charMap[currChar] = i
		maxLen = max(maxLen, i-left+1)
	}
	return maxLen
}

// 25. K 个一组翻转链表
func reverseKGroup(head *ListNode, k int) *ListNode {
	if k <= 1 {
		return head
	}
	zeroNode := &ListNode{Next: head}
	p0 := zeroNode

	for {
		// 检查剩余节点是否至少有k个
		curr := p0.Next
		for i := 0; i < k; i++ {
			if curr == nil {
				return zeroNode.Next // 不足k个，直接返回
			}
			curr = curr.Next
		}

		// 翻转当前k个节点
		start := p0.Next
		pre := p0
		curr = start
		for i := 0; i < k; i++ {
			next := curr.Next
			curr.Next = pre
			pre = curr
			curr = next
		}

		// 连接翻转后的链表
		p0.Next = pre
		start.Next = curr
		p0 = start
	}
}
