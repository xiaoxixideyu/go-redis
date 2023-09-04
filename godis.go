package main

import (
	"errors"
	"hash/fnv"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type CmdType byte

const (
	COMMAND_UNKNOW CmdType = 0x00
	COMMAND_INLINE CmdType = 0x01
	COMMAND_BULK   CmdType = 0x02
)

const (
	GODIS_IO_BUF     int = 1024 * 16
	GODIS_MAX_BULK   int = 1024 * 4
	GODIS_MAX_INLINE int = 1024 * 4
)

const EXPIRE_CHECK_COUNT int = 100

type GodisDB struct {
	data   *Dict
	expire *Dict
}

type GodisServer struct {
	fd      int
	port    int
	db      *GodisDB
	clients map[int]*GodisClient
	aeLoop  *AeLoop
}

type GodisClient struct {
	fd       int
	db       *GodisDB
	args     []*Gobj
	reply    *List
	sentLen  int
	queryBuf []byte
	queryLen int
	cmdTy    CmdType
	bulkNum  int
	bulkLen  int
}

type CommandProc func(c *GodisClient)

type GodisCommand struct {
	name  string
	proc  CommandProc
	arity int
}

var server GodisServer
var cmdTable []GodisCommand = []GodisCommand{
	{"get", GetCommand, 2},
	{"set", SetCommand, 3},
	{"strlen", StrlenCommand, 2},
	{"incr", IncrCommand, 2},
	{"incrby", IncrByCommand, 3},
	{"decr", DecrCommand, 2},
	{"decrby", DecrByCommand, 3},
	{"del", DelCommand, 2},
	{"expire", ExpireCommand, 3},
	{"ttl", TtlCommand, 2},
}

func CreateClient(fd int) *GodisClient {
	return &GodisClient{
		fd:       fd,
		db:       server.db,
		queryBuf: make([]byte, GODIS_IO_BUF),
		reply:    CreateList(ListType{EqualFunc: GStrEqual}),
	}
}

func freeClient(client *GodisClient) {
	freeArgs(client)
	delete(server.clients, client.fd)
	server.aeLoop.RemoveFileEvent(client.fd, AE_READABLE)
	server.aeLoop.RemoveFileEvent(client.fd, AE_WRITABLE)
	freeReplyList(client)
	Close(client.fd)
}

func freeArgs(client *GodisClient) {
	for _, arg := range client.args {
		arg.DecrRefCount()
	}
}

func freeReplyList(client *GodisClient) {
	for client.reply.Length() != 0 {
		n := client.reply.head
		client.reply.DelNode(n)
		n.Val.DecrRefCount()
	}
}

func resetClient(client *GodisClient) {
	freeArgs(client)
	client.cmdTy = COMMAND_UNKNOW
	client.bulkLen = 0
	client.bulkNum = 0
}

func expireIfNeeded(key *Gobj) {
	entry := server.db.expire.Find(key)
	if entry == nil {
		return
	}
	when := entry.Val.IntVal()
	if when > GetMsTime() {
		return
	}
	server.db.data.Delete(key)
	server.db.expire.Delete(key)
}

func findKeyRead(key *Gobj) *Gobj {
	expireIfNeeded(key)
	return server.db.data.Get(key)
}

func (client *GodisClient) findLineInQuery() (int, error) {
	index := strings.Index(string(client.queryBuf[:client.queryLen]), "\r\n")
	if index < 0 && client.queryLen > GODIS_MAX_INLINE {
		return index, errors.New("too big inline cmd")
	}
	return index, nil
}

func (client *GodisClient) getNumInQuery(start, end int) (int, error) {
	num, err := strconv.Atoi(string(client.queryBuf[start:end]))
	client.queryBuf = client.queryBuf[end+2:]
	client.queryLen -= end + 2
	return num, err
}

func lookupCommand(cmdStr string) *GodisCommand {
	for _, c := range cmdTable {
		if strings.EqualFold(cmdStr, c.name) {
			return &c
		}
	}
	return nil
}

func ProcessCommand(client *GodisClient) {
	cmdStr := client.args[0].StrVal()
	log.Printf("process command: %v\n", cmdStr)
	if cmdStr == "quit" {
		freeClient(client)
		return
	}
	cmd := lookupCommand(cmdStr)
	if cmd == nil {
		client.AddReplyStr("-ERR: unknow commandr\r\n")
		resetClient(client)
	} else if cmd.arity != len(client.args) {
		client.AddReplyStr("-ERR: wrong number of args\r\n")
		resetClient(client)
	} else {
		cmd.proc(client)
		resetClient(client)
	}
}

func handleInlineBuf(client *GodisClient) (bool, error) {
	index, err := client.findLineInQuery()
	if index < 0 {
		return false, err
	}

	str := string(client.queryBuf[:index])
	subs := strings.Split(str, " ")
	client.queryBuf = client.queryBuf[index+2:]
	client.queryLen -= index + 2
	client.args = make([]*Gobj, len(subs))
	for i, v := range subs {
		client.args[i] = CreateObject(GSTR, v)
	}
	return true, nil
}

func handleBulkBuf(client *GodisClient) (bool, error) {
	if client.bulkNum == 0 {
		index, err := client.findLineInQuery()
		if index < 0 {
			return false, err
		}
		bulkNum, err := client.getNumInQuery(1, index)
		if err != nil {
			return false, err
		}
		client.bulkNum = bulkNum
		client.args = make([]*Gobj, bulkNum)
	}
	for client.bulkNum > 0 {
		if client.bulkLen == 0 {
			index, err := client.findLineInQuery()
			if index < 0 {
				return false, err
			}
			if client.queryBuf[0] != '$' {
				return false, errors.New("expect $ for bulk length")
			}
			blen, err := client.getNumInQuery(1, index)
			if err != nil || blen == 0 {
				return false, err
			}
			if blen > GODIS_MAX_BULK {
				return false, errors.New("too big bulk")
			}
			client.bulkLen = blen
		}
		if client.queryLen < client.bulkLen+2 {
			return false, nil
		}
		index := client.bulkLen
		if client.queryBuf[index] != '\r' || client.queryBuf[index+1] != '\n' {
			return false, errors.New("expect CRLF for bulk end")
		}
		client.args[len(client.args)-client.bulkNum] = CreateObject(GSTR, string(client.queryBuf[:index]))
		client.queryBuf = client.queryBuf[index+2:]
		client.queryLen -= index + 2
		client.bulkLen = 0
		client.bulkNum--
	}
	return true, nil
}

func ProcessQueryBuf(client *GodisClient) error {
	for client.queryLen > 0 {
		if client.cmdTy == COMMAND_UNKNOW {
			if client.queryBuf[0] == '*' {
				client.cmdTy = COMMAND_BULK
			} else {
				client.cmdTy = COMMAND_INLINE
			}
		}
		var ok bool
		var err error
		if client.cmdTy == COMMAND_INLINE {
			ok, err = handleInlineBuf(client)
		} else if client.cmdTy == COMMAND_BULK {
			ok, err = handleBulkBuf(client)
		} else {
			return errors.New("unknow Godis Command Type")
		}
		if err != nil {
			return err
		}
		if ok {
			if len(client.args) == 0 {
				resetClient(client)
			} else {
				ProcessCommand(client)
			}
		} else {
			break
		}
	}
	return nil
}

func ReadQueryFromClient(loop *AeLoop, fd int, extra interface{}) {
	client := extra.(*GodisClient)
	if len(client.queryBuf)-client.queryLen < GODIS_MAX_BULK {
		client.queryBuf = append(client.queryBuf, make([]byte, GODIS_MAX_BULK)...)
	}
	n, err := Read(fd, client.queryBuf[client.queryLen:])
	if err != nil {
		log.Printf("client %v read err: %v\n", fd, err)
		freeClient(client)
		return
	}
	client.queryLen += n
	log.Printf("read %v bytes from client: %v\n", n, client.fd)
	if err = ProcessQueryBuf(client); err != nil {
		log.Printf("process query buf err: %v\n", err)
		freeClient(client)
	}
}

func SendReplyToClient(loop *AeLoop, fd int, extra interface{}) {
	client := extra.(*GodisClient)
	for client.reply.Length() > 0 {
		rep := client.reply.First()
		buf := []byte(rep.Val.StrVal())
		bufLen := len(buf)
		if client.sentLen < bufLen {
			n, err := Write(client.fd, buf[client.sentLen:])
			if err != nil {
				log.Printf("send reply err: %v\n", err)
				freeClient(client)
				return
			}
			client.sentLen += n
			log.Printf("send %v bytes to client:%v\n", n, client.fd)
			if client.sentLen == bufLen {
				client.reply.DelNode(rep)
				rep.Val.DecrRefCount()
				client.sentLen = 0
			} else {
				break
			}
		}
	}
	if client.reply.Length() == 0 {
		loop.RemoveFileEvent(client.fd, AE_WRITABLE)
		client.sentLen = 0
	}
}

func (client *GodisClient) AddReply(o *Gobj) {
	client.reply.RPush(o)
	o.IncrRefCount()
	server.aeLoop.AddFileEvent(client.fd, AE_WRITABLE, SendReplyToClient, client)
}

func (client *GodisClient) AddReplyStr(str string) {
	o := CreateObject(GSTR, str)
	client.AddReply(o)
	o.DecrRefCount()
}

func AcceptHandler(loop *AeLoop, fd int, extra interface{}) {
	cfd, err := Accept(fd)
	if err != nil {
		log.Printf("accept err: %v\n", err)
		return
	}
	client := CreateClient(cfd)
	server.clients[cfd] = client
	loop.AddFileEvent(cfd, AE_READABLE, ReadQueryFromClient, client)
	log.Printf("accept client, fd: %v\n", cfd)
}

func ServerCron(loop *AeLoop, id int, extra interface{}) {
	for i := 0; i < EXPIRE_CHECK_COUNT; i++ {
		entry := server.db.expire.RandomGet()
		if entry == nil {
			break
		}
		if entry.Val.IntVal() < time.Now().Unix() {
			server.db.data.Delete(entry.Key)
			server.db.expire.Delete(entry.Key)
		}
	}
}

func GStrEqual(a, b *Gobj) bool {
	if a.Type_ != GSTR || b.Type_ != GSTR {
		return false
	}
	return a.StrVal() == b.StrVal()
}

func GStrHash(key *Gobj) int64 {
	if key.Type_ != GSTR {
		return 0
	}
	hash := fnv.New64a()
	hash.Write([]byte(key.StrVal()))
	return int64(hash.Sum64())
}

func initServer(config *Config) error {
	server.port = config.Port
	server.db = &GodisDB{
		data:   DictCreate(DictType{HashFunc: GStrHash, EqualFunc: GStrEqual}),
		expire: DictCreate(DictType{HashFunc: GStrHash, EqualFunc: GStrEqual}),
	}
	server.clients = make(map[int]*GodisClient)
	var err error
	if server.aeLoop, err = AeLoopCreate(); err != nil {
		return err
	}
	server.fd, err = TCPServer(server.port)
	return err
}

func main() {
	path := os.Args[1]
	config, err := LoadConfig(path)
	if err != nil {
		log.Printf("config error: %v\n", err)
	}
	err = initServer(config)
	if err != nil {
		log.Printf("init server error: %v\n", err)
		return
	}
	server.aeLoop.AddFileEvent(server.fd, AE_READABLE, AcceptHandler, nil)
	server.aeLoop.AddTimeEvent(AE_NORMAL, 100, ServerCron, nil)
	log.Printf("godis server is up")
	server.aeLoop.AeMain()
}
