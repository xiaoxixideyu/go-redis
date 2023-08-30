package main

import "strconv"

type GType = uint8

const (
	GSTR  GType = 0x00
	GLIST GType = 0x01
	GSET  GType = 0x02
	GZSET GType = 0x03
	GDICT GType = 0x04
)

type GVal interface{}

type Gobj struct {
	Type_    GType
	Val_     GVal
	refCount int
}

func (o *Gobj) IntVal() int64 {
	if o.Type_ != GSTR {
		return 0
	}
	val, _ := strconv.ParseInt(o.Val_.(string), 10, 64)
	return val
}

func (o *Gobj) StrVal() string {
	if o.Type_ != GSTR {
		return ""
	}
	return o.Val_.(string)
}

func CreateFromInt(val int64) *Gobj {
	return &Gobj{
		Type_:    GSTR,
		Val_:     strconv.FormatInt(val, 10),
		refCount: 1,
	}
}

func CreateObject(typ GType, ptr interface{}) *Gobj {
	return &Gobj{
		Type_:    typ,
		Val_:     ptr,
		refCount: 1,
	}
}

func (o *Gobj) IncrRefCount() {
	o.refCount++
}

func (o *Gobj) DecrRefCount() {
	o.refCount--
	if o.refCount == 0 {
		o.Val_ = nil
	}
}
