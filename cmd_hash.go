package main

import "fmt"

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
