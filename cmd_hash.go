package main

import (
	"fmt"
	"strconv"
)

func HSetCommand(client *GodisClient) {
	setProcess(client)
}

func HMSetCommand(client *GodisClient) {
	setProcess(client)
}

func HGetCommand(client *GodisClient) {
	getProcess(client)
}

func HMGetCommand(client *GodisClient) {
	getProcess(client)
}

func setProcess(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		newDict := DictCreate(DictType{GStrHash, GStrEqual})
		newObj := CreateObject(GDICT, newDict)
		defer newObj.DecrRefCount()
		client.db.data.Set(key, newObj)
		val = newObj
	} else if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	for i := 2; i < len(client.args); i += 2 {
		hashKey := client.args[i]
		hashVal := client.args[i+1]
		dict := val.DictVal()
		dict.Set(hashKey, hashVal)
	}

	client.AddReplyStr("+OK\r\n")
}

func getProcess(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		for i := 2; i < len(client.args); i++ {
			client.AddReplyStr("$-1\r\n")
		}
		return
	} else if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	dict := val.DictVal()
	for i := 2; i < len(client.args); i++ {
		hashKey := client.args[i]
		hashVal := dict.Get(hashKey)
		if hashVal == nil {
			client.AddReplyStr("$-1\r\n")
		} else {
			str := hashVal.StrVal()
			client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(str), str))
		}
	}
}

func HLenCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
		return
	}
	if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	dict := val.DictVal()
	length := dict.GetUsed()
	str := strconv.FormatInt(length, 10)
	client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(str), str))
}

func HKeysCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
		return
	}
	if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	dict := val.DictVal()
	keys := dict.GetAllKey()
	if len(keys) == 0 {
		client.AddReplyStr("$-1\r\n")
	} else {
		for _, hashKey := range keys {
			client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(hashKey), hashKey))
		}
	}
}

func HIncrByCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
		return
	}
	if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	hashKey := client.args[2]
	incrNum, err := strconv.ParseInt(client.args[3].StrVal(), 10, 64)
	if err != nil {
		client.AddReplyStr("-ERR: wrong type of incr number\r\n")
		return
	}

	dict := val.DictVal()
	hashVal := dict.Get(hashKey)
	if hashVal == nil {
		client.AddReplyStr("$-1\r\n")
		return
	}

	valNum, err := strconv.ParseInt(hashVal.StrVal(), 10, 64)
	if err != nil {
		client.AddReplyStr("-ERR: Not a usable number\r\n")
		return
	}

	valNum += incrNum
	newVal := CreateFromInt(valNum)
	defer newVal.DecrRefCount()
	dict.Set(hashKey, newVal)
	str := newVal.StrVal()
	client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(str), str))
}

func HGetAllCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
		return
	}
	if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	dict := val.DictVal()
	entries := dict.GetAllEntry()
	if len(entries) == 0 {
		client.AddReplyStr("$-1\r\n")
	} else {
		for _, e := range entries {
			hashKey := e.Key.StrVal()
			hashVal := e.Val.StrVal()
			client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(hashKey), hashKey))
			client.AddReplyStr(fmt.Sprintf("$%d %s\r\n", len(hashVal), hashVal))
		}
	}
}

func HExistsCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
		return
	}
	if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	dict := val.DictVal()
	hashKey := client.args[2]
	hashVal := dict.Get(hashKey)
	if hashVal == nil {
		client.AddReplyStr("$-1\r\n")
	} else {
		client.AddReplyStr("+OK\r\n")
	}
}

func HSetNxCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		newDict := DictCreate(DictType{GStrHash, GStrEqual})
		newObj := CreateObject(GDICT, newDict)
		defer newObj.DecrRefCount()
		client.db.data.Set(key, newObj)
		val = newObj
	} else if val.Type_ != GDICT {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}

	hashKey := client.args[2]
	hashVal := client.args[3]
	dict := val.DictVal()
	currVal := dict.Get(hashKey)
	if currVal == nil {
		dict.Set(hashKey, hashVal)
		client.AddReplyStr("+OK\r\n")
	} else {
		client.AddReplyStr("$-1\r\n")
	}
}
