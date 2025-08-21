package suanfa

import (
	"fmt"
	"testing"
)

//146. LRU 缓存

func TestSuanfa003(t *testing.T) {
	lru := Constructor(2)
	lru.Put(1, 1)           // 缓存: {1=1}
	lru.Put(2, 2)           // 缓存: {1=1, 2=2}
	fmt.Println(lru.Get(1)) // 返回 1，缓存: {2=2, 1=1}
	lru.Put(3, 3)           // 淘汰 2，缓存: {1=1, 3=3}
	fmt.Println(lru.Get(2)) // 返回 -1 (未找到)
	lru.Put(4, 4)           // 淘汰 1，缓存: {3=3, 4=4}
	fmt.Println(lru.Get(1)) // 返回 -1 (未找到)
	fmt.Println(lru.Get(3)) // 返回 3
	fmt.Println(lru.Get(4)) // 返回 4
}

type LRUNode struct {
	key   int
	value int
	prev  *LRUNode
	next  *LRUNode
}

type LRUCache struct {
	cacheSize int
	cache     map[int]*LRUNode
	head      *LRUNode
	tail      *LRUNode
}

func Constructor(capacity int) LRUCache {
	head := &LRUNode{}
	tail := &LRUNode{}
	head.next = tail
	tail.prev = head
	return LRUCache{
		cacheSize: capacity,
		cache:     make(map[int]*LRUNode),
		head:      head,
		tail:      tail,
	}
}

func (this *LRUCache) Get(key int) int {
	// 从缓存中查找节点，ok表示是否找到
	node, ok := this.cache[key]
	if !ok {
		// 如果未找到，返回-1
		return -1
	}
	// 如果找到，将该节点移动到链表头部（表示最近使用）
	this.moveToHead(node)
	// 返回节点的值
	return node.value
}

// 缓存中添加或者更新值
func (this *LRUCache) Put(key int, value int) {
	node, ok := this.cache[key]
	if ok {
		node.value = value
		this.moveToHead(node)
		return
	}
	newNode := &LRUNode{
		key:   key,
		value: value,
	}
	this.cache[key] = newNode
	this.addToHead(newNode)

	if len(this.cache) > this.cacheSize {
		removedNode := this.removeTail()
		delete(this.cache, removedNode.key)
	}

}

func (this *LRUCache) moveToHead(node *LRUNode) {
	this.removeNode(node)
	this.addToHead(node)
}

func (this *LRUCache) removeNode(node *LRUNode) {
	//跳过我
	node.prev.next = node.next // 前节点的next 指向我的next
	node.next.prev = node.prev //
}

func (this *LRUCache) addToHead(node *LRUNode) {
	node.prev = this.head
	node.next = this.head.next
	this.head.next.prev = node
	this.head.next = node
}

func (this *LRUCache) removeTail() *LRUNode {
	node := this.tail.prev
	this.removeNode(node)
	return node
}

/**
 * Your LRUCache object will be instantiated and called as such:
 * obj := Constructor(capacity);
 * param_1 := obj.Get(key);
 * obj.Put(key,value);
 */
