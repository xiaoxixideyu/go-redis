package main

func ExpireCommand(client *GodisClient) {
	key := client.args[1]
	val := client.args[2]
	if val.Type_ != GSTR {
		client.AddReplyStr("-ERR: wrong type\r\n")
		return
	}
	expire := GetMsTime() + (val.IntVal() * 1000)
	expObj := CreateFromInt(expire)
	server.db.expire.Set(key, expObj)
	expObj.DecrRefCount()
	client.AddReplyStr("+OK\r\n")
}

func DelCommand(client *GodisClient) {
	key := client.args[1]
	val := findKeyRead(key)
	if val == nil {
		client.AddReplyStr("$-1\r\n")
	} else {
		client.db.data.Delete(key)
		server.db.expire.Delete(key)
		client.AddReplyStr("+OK\r\n")
	}
}
