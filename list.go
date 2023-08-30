package main

type List struct {
	head   *Node
	tail   *Node
	length int
}

type Node struct {
	Val  *Gobj
	prev *Node
	next *Node
}

func (list *List) EqualFunc(a, b *Gobj) bool {
	return a == b
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
	if list.length == 0 {
		list.head = &n
		list.tail = &n
	} else {
		n.prev = list.tail
		list.tail.next = &n
		list.tail = &n
	}
	list.length++
}

func (list *List) DelNode(n *Node) {
	if n == nil {
		return
	}
	if n == list.head {
		n.next.prev = nil
		list.head = n.next
		n.next = nil
	} else if n == list.tail {
		n.prev.next = nil
		list.tail = n.prev
		n.prev = nil
	} else {
		n.next.prev = n.prev
		n.prev.next = n.next
		n.next = nil
		n.prev = nil
	}
	list.length--
}

func (list *List) Delete(val *Gobj) {
	list.DelNode(list.Find(val))
}
