package main

import (
	"fmt"
	"strconv"
)

func GetCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
	} else if val.Type_ != GSTR {
		client.AddReplyStr("-ERR: wrong type\r\n")
	} else {
		str := val.StrVal()
		client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(str), str))
	}
}

func SetCommand(client *GodisClient) {
	key := client.args[1]
	val := client.args[2]
	if val.Type_ != GSTR {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}
	server.db.data.Set(key, val)
	server.db.expire.Delete(key)
	client.AddReplyStr("+OK\r\n")
}

func StrlenCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
	} else if val.Type_ != GSTR {
		client.AddReplyStr("-ERR: wrong type\r\n")
	} else {
		str := strconv.Itoa(len(val.StrVal()))
		client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(str), str))
	}
}

func IncrCommand(client *GodisClient) {
	NumberProcess(client, 1)
}

func IncrByCommand(client *GodisClient) {
	incrNum, err := strconv.ParseInt(client.args[2].StrVal(), 10, 64)
	if err != nil {
		client.AddReplyStr("-ERR: wrong type of args")
	} else {
		NumberProcess(client, incrNum)
	}
}

func DecrCommand(client *GodisClient) {
	NumberProcess(client, -1)
}

func DecrByCommand(client *GodisClient) {
	decrNum, err := strconv.ParseInt(client.args[2].StrVal(), 10, 64)
	if err != nil {
		client.AddReplyStr("-ERR: wrong type of args")
	} else {
		NumberProcess(client, decrNum*-1)
	}
}

func NumberProcess(client *GodisClient, diff int64) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
	} else if val.Type_ != GSTR {
		client.AddReplyStr("-ERR: wrong type\r\n")
	} else if num, err := strconv.ParseInt(val.StrVal(), 10, 64); err == nil {
		num += diff
		newVal := CreateFromInt(num)
		client.db.data.Set(key, newVal)
		str := newVal.StrVal()
		client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(str), str))
	} else {
		client.AddReplyStr("-ERR: Not a usable number\r\n")
	}
}
