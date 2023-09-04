package main

import (
	"math/rand"
	"time"
)

type Node struct {
	Val  *Gobj
	prev *Node
	next *Node
}

type ListType struct {
	EqualFunc func(a, b *Gobj) bool
}

type List struct {
	ListType
	head   *Node
	tail   *Node
	length int
}

func (list *List) EqualFunc(a, b *Gobj) bool {
	return a == b
}

func CreateList(listType ListType) *List {
	return &List{ListType: listType}
}

func (list *List) Length() int {
	return list.length
}

func (list *List) First() *Node {
	return list.head
}

func (list *List) Last() *Node {
	return list.tail
}

func (list *List) Find(val *Gobj) *Node {
	p := list.head
	for p != nil {
		if list.EqualFunc(p.Val, val) {
			break
		}
		p = p.next
	}
	return p
}

func (list *List) LPush(val *Gobj) {
	var n Node
	n.Val = val
	val.IncrRefCount()
	if list.length == 0 {
		list.head = &n
		list.tail = &n
	} else {
		n.next = list.head
		list.head.prev = &n
		list.head = &n
	}
	list.length++
}

func (list *List) RPush(val *Gobj) {
	var n Node
	n.Val = val
	val.IncrRefCount()
	if list.length == 0 {
		list.head = &n
		list.tail = &n
	} else {
		n.prev = list.tail
		list.tail.next = &n
		list.tail = &n
	}
	list.length += 1
}

func (list *List) LRange(left, right int) *Node {
	if list.Length() == 0 || left >= list.length || right < 0 {
		return nil
	}
	if left < 0 {
		left = 0
	}
	if right >= list.length {
		right = list.length - 1
	}
	mod := right - left + 1
	newRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := left + newRand.Intn(mod)
	n := list.head
	for index != 0 {
		n = n.next
		index--
	}
	return n
}

func (list *List) DelNode(n *Node) {
	if n == nil {
		return
	}
	if n == list.head {
		if n.next != nil {
			n.next.prev = nil
		}
		list.head = n.next
		if list.length == 1 {
			list.tail = nil
		}
		n.next = nil
	} else if n == list.tail {
		if n.prev != nil {
			n.prev.next = nil
		}
		list.tail = n.prev
		n.prev = nil
	} else {
		if n.next != nil {
			n.next.prev = n.prev
		}
		if n.prev != nil {
			n.prev.next = n.next
		}
		n.next = nil
		n.prev = nil
	}
	n.Val.DecrRefCount()
	list.length -= 1
}

func (list *List) Delete(val *Gobj) {
	list.DelNode(list.Find(val))
}
