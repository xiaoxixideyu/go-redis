package main

import (
	"errors"
	"math"
	"math/rand"
)

const (
	INIT_SIZE    int64 = 8
	FORCE_RATIO  int64 = 2
	GROW_RATIO   int64 = 2
	DEFAULT_STEP int   = 1
)

var (
	EP_ERR = errors.New("expend error")
	EX_ERR = errors.New("keys exists error")
	NK_ERR = errors.New("key doesn't exist error")
)

type Entry struct {
	Key  *Gobj
	Val  *Gobj
	next *Entry
}

type htable struct {
	table []*Entry
	size  int64
	mask  int64
	used  int64
}

type DictType struct {
	HashFunc  func(key *Gobj) int64
	EqualFunc func(key1, key2 *Gobj) bool
}

type Dict struct {
	DictType
	hts         [2]*htable
	rehashIndex int64
}

func DictCreate(dictType DictType) *Dict {
	return &Dict{
		DictType:    dictType,
		rehashIndex: -1,
	}
}

func (dict *Dict) isRehashing() bool {
	return dict.rehashIndex != -1
}

func (dict *Dict) rehashStep() {
	dict.rehash(DEFAULT_STEP)
}

func (dict *Dict) rehash(step int) {
	for step > 0 {
		if dict.hts[0].used == 0 {
			dict.hts[0] = dict.hts[1]
			dict.hts[1] = nil
			dict.rehashIndex = -1
			return
		}
		for dict.hts[0].table[dict.rehashIndex] == nil {
			dict.rehashIndex++
		}
		entry := dict.hts[0].table[dict.rehashIndex]
		for entry != nil {
			ne := entry.next
			idx := dict.HashFunc(entry.Key) & dict.hts[1].mask
			entry.next = dict.hts[1].table[idx]
			dict.hts[1].table[idx] = entry
			dict.hts[0].used--
			dict.hts[1].used++
			entry = ne
		}
		dict.hts[0].table[dict.rehashIndex] = nil
		dict.rehashIndex++
		step--
	}
}

func nextPower(size int64) int64 {
	for i := INIT_SIZE; i < math.MaxInt64; i *= 2 {
		if i >= size {
			return i
		}
	}
	return -1
}

func (dict *Dict) expand(size int64) error {
	sz := nextPower(size)
	if dict.isRehashing() || (dict.hts[0] != nil && dict.hts[0].size >= sz) {
		return EP_ERR
	}
	var ht htable
	ht.table = make([]*Entry, sz)
	ht.size = sz
	ht.mask = sz - 1
	ht.used = 0
	if dict.hts[0] == nil {
		dict.hts[0] = &ht
		return nil
	}
	dict.hts[1] = &ht
	dict.rehashIndex = 0
	return nil
}

func (dict *Dict) expandIfNeed() error {
	if dict.isRehashing() {
		return nil
	}
	if dict.hts[0] == nil {
		dict.expand(INIT_SIZE)
		return nil
	}
	if (dict.hts[0].used > dict.hts[0].size) && (dict.hts[0].used/dict.hts[0].size > FORCE_RATIO) {
		return dict.expand(dict.hts[0].size * GROW_RATIO)
	}
	return nil
}

func (dict *Dict) keyIndex(key *Gobj) int64 {
	err := dict.expandIfNeed()
	if err != nil {
		return -1
	}
	h := dict.HashFunc(key)
	var idx int64
	for i := 0; i <= 1; i++ {
		idx = h & dict.hts[i].mask
		entry := dict.hts[i].table[idx]
		for entry != nil {
			if dict.EqualFunc(key, entry.Key) {
				return -1
			}
			entry = entry.next
		}
		if !dict.isRehashing() {
			break
		}
	}
	return idx
}

func (dict *Dict) AddRaw(key *Gobj) *Entry {
	if dict.isRehashing() {
		dict.rehashStep()
	}
	idx := dict.keyIndex(key)
	if idx == -1 {
		return nil
	}
	var ht *htable
	if dict.isRehashing() {
		ht = dict.hts[1]
	} else {
		ht = dict.hts[0]
	}
	var entry Entry
	entry.Key = key
	key.IncrRefCount()
	entry.next = ht.table[idx]
	ht.table[idx] = &entry
	ht.used++
	return &entry
}

func (dict *Dict) Add(key, val *Gobj) error {
	entry := dict.AddRaw(key)
	if entry == nil {
		return EX_ERR
	}
	entry.Val = val
	val.IncrRefCount()
	return nil
}

func (dict *Dict) Set(key, val *Gobj) {
	if err := dict.Add(key, val); err == nil {
		return
	}
	entry := dict.Find(key)
	entry.Val.DecrRefCount()
	entry.Val = val
	val.IncrRefCount()
}

func freeEntry(e *Entry) {
	e.Key.DecrRefCount()
	e.Val.DecrRefCount()
	e.next = nil
}

func (dict *Dict) Delete(key *Gobj) error {
	if dict.hts[0] == nil {
		return NK_ERR
	}
	if dict.isRehashing() {
		dict.rehashStep()
	}
	h := dict.HashFunc(key)
	for i := 0; i <= 1; i++ {
		idx := h & dict.hts[i].mask
		entry := dict.hts[i].table[idx]
		var prev *Entry
		for entry != nil {
			if dict.EqualFunc(key, entry.Key) {
				if prev == nil {
					dict.hts[i].table[idx] = entry.next
				} else {
					prev.next = entry.next
				}
				freeEntry(entry)
				return nil
			}
			prev = entry
			entry = entry.next
		}
		if !dict.isRehashing() {
			break
		}
	}
	return NK_ERR
}

func (dict *Dict) Find(key *Gobj) *Entry {
	if dict.hts[0] == nil {
		return nil
	}
	if dict.isRehashing() {
		dict.rehashStep()
	}
	h := dict.HashFunc(key)
	for i := 0; i <= 1; i++ {
		idx := h & dict.hts[i].mask
		entry := dict.hts[i].table[idx]
		for entry != nil {
			if dict.EqualFunc(key, entry.Key) {
				return entry
			}
			entry = entry.next
		}
		if !dict.isRehashing() {
			break
		}
	}
	return nil
}

func (dict *Dict) Get(key *Gobj) *Gobj {
	entry := dict.Find(key)
	if entry == nil {
		return nil
	}
	return entry.Val
}

func (dict *Dict) RandomGet() *Entry {
	if dict.hts[0] == nil {
		return nil
	}
	t := dict.hts[0]
	if dict.isRehashing() {
		dict.rehashStep()
		if dict.hts[1] != nil && dict.hts[1].used > dict.hts[0].used {
			t = dict.hts[1]
		}
	}
	idx := rand.Int63n(t.size)
	cnt := 0
	for t.table[idx] == nil && cnt < 1000 {
		idx = rand.Int63n(t.size)
		cnt++
	}
	if t.table[idx] == nil {
		return nil
	}
	listLen := int64(0)
	entry := t.table[idx]
	for entry != nil {
		listLen++
		entry = entry.next
	}
	listIdx := rand.Int63n(listLen)
	entry = t.table[idx]
	for listIdx > 0 {
		entry = entry.next
		listIdx--
	}
	return entry
}
