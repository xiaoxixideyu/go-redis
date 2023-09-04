package main

import (
	"fmt"
)

func LPushCommand(client *GodisClient) {
	pushProcess(client, true)
}

func RPushCommand(client *GodisClient) {
	pushProcess(client, false)
}

func LPopCommand(client *GodisClient) {
	popProcess(client, true)
}

func RPopCommand(client *GodisClient) {
	popProcess(client, false)
}

func pushProcess(client *GodisClient, left bool) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		newList := CreateList(ListType{EqualFunc: GStrEqual})
		newVal := CreateObject(GLIST, newList)
		client.db.data.Set(key, newVal)
		val = newVal
		newVal.DecrRefCount()
	} else if val.Type_ != GLIST {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}
	list := val.ListVal()
	for i := 2; i < len(client.args); i++ {
		if left {
			list.LPush(client.args[i])
		} else {
			list.RPush(client.args[i])
		}
	}
	client.AddReplyStr("+OK\r\n")
}

func popProcess(client *GodisClient, left bool) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
		return
	}
	if val.Type_ != GLIST {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}
	list := val.ListVal()
	if list.Length() == 0 {
		client.AddReplyStr("$-1\r\n")
		return
	}
	var head *Node
	if left {
		head = list.First()
	} else {
		head = list.Last()
	}
	headVal := head.Val
	if headVal.Type_ != GSTR {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}
	headValStr := headVal.StrVal()
	client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(headValStr), headValStr))
	list.DelNode(head)
}
