package main

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
	list.length += 1
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
	list.length -= 1
}

func (list *List) Delete(val *Gobj) {
	list.DelNode(list.Find(val))
}
