package command

import (
	"bytes"
	"fmt"
	"github.com/meitu/titan/context"
	"github.com/meitu/titan/encoding/resp"
	"github.com/yuin/gopher-lua"
	"strconv"
	"strings"
	"time"
)

var cli *context.ClientContext

func NewClientContext(ctx *Context) *context.ClientContext {
	now := time.Now()
	cli := &context.ClientContext{
		Namespace:  ctx.Client.Namespace,
		ID:         ctx.Client.ID,
		RemoteAddr: ctx.Client.RemoteAddr,
		DB:         ctx.Client.DB,
		Created:    now,
		Updated:    now,
	}
	return cli
}

const luaRedisTypeName = "redis"

// Registers my redis type to given L.
func registerRedisType(L *lua.LState) {
	//
	mt := L.NewTypeMetatable(luaRedisTypeName)
	L.SetGlobal("redis", mt)
	// static attributes
	L.SetField(mt, "call", L.NewFunction(newRedisCall))
}

func parseLuaValue(ctx *Context, data lua.LValue) {
	Type := data.Type().String()
	if Type == "string" {
		if lv, ok := data.(lua.LString); ok {
			resp.ReplySimpleString(ctx.Out, string(lv))
		}
	} else if Type == "number" {
		if intv, ok := data.(lua.LNumber); ok {
			resp.ReplyInteger(ctx.Out, int64(intv))
		}
	} else if Type == "boolean" {
		resp.ReplyNullBulkString(ctx.Out)
	} else if Type == "nil" {
		resp.ReplyNullBulkString(ctx.Out)
	} else if Type == "table" {
		var array []lua.LValue
		a := data.(*lua.LTable)
		len := 0
		a.ForEach(func(value lua.LValue, value2 lua.LValue) {
			len++
			array = append(array, value2)
		})
		resp.ReplyArray(ctx.Out, len)
		for _, r := range array {
			parseLuaValue(ctx, r)
		}
	} else {
		resp.ReplyNullBulkString(ctx.Out)
	}
}

// Constructor
func newRedisCall(L *lua.LState) int {
	var rest []string
	// filter data from lua to redis command without first arg
	GetLuaTop := L.GetTop()
	for i := 2; i <= GetLuaTop; i++ {
		lv := L.Get(i)
		rest = append(rest, lv.String())
	}

	// redis call command
	serv := &context.ServerContext{}
	serv.Clients.Store(cli.RemoteAddr, cli)
	out := bytes.NewBuffer(nil)
	ctx := &Context{
		Name:    L.Get(1).String(),
		Args:    rest,
		In:      nil,
		Out:     out,
		Context: context.New(cli, serv),
	}

	Call(ctx)
	data := out.String()

	//return array of string
	dArray := resp.NewDecoder(bytes.NewBufferString(data))
	valArray, errArray := dArray.Array()
	if errArray == nil {
		for i := 1; i <= valArray; i++ {
			valBulkString, _ := dArray.BulkString()
			L.Push(lua.LString(valBulkString))
		}
		return valArray
	}

	//return only on string
	dSimpleString := resp.NewDecoder(bytes.NewBufferString(data))
	valSimpleString, errSimpleString := dSimpleString.SimpleString()
	if errSimpleString == nil {
		L.Push(lua.LString(valSimpleString))
		return 1
	}

	//return only on string
	dBulkString := resp.NewDecoder(bytes.NewBufferString(data))
	valBulkString, errBulkString := dBulkString.BulkString()
	if errBulkString == nil {
		L.Push(lua.LString(valBulkString))
		return 1
	}

	//return integer
	dInteger := resp.NewDecoder(bytes.NewBufferString(data))
	valInteger, errInteger := dInteger.Integer()
	if errInteger == nil {
		L.Push(lua.LNumber(valInteger))
		return 1
	}

	return 0
}

// Eval lua script the server
func Eval(ctx *Context) {

	args := ctx.Args
	if len(args) < 2 {
		resp.ReplyError(ctx.Out, ErrWrongArgs(ctx.Name).Error())
		return
	}

	var (
		err     error
		keysLen int
	)

	keysLen, err = strconv.Atoi(string(args[1]))
	if err != nil {
		resp.ReplyError(ctx.Out, ErrWrongArgs(ctx.Name).Error())
		return
	}
	keysLen = keysLen + 2
	// check len KEYS more than all c.args without lua-script(c.args[0]) and keysLen(c.args[1])
	if keysLen > len(args) {
		resp.ReplyError(ctx.Out, ErrWrongArgs(ctx.Name).Error())
		return
	}

	luaScript := fmt.Sprintf(args[0][:])
	// replace KEYS for lua
	keysArray := make([]string, len(args[2:keysLen]))
	if keysLen > 2 {
		for key := range keysArray {
			luaScript = strings.Replace(luaScript, fmt.Sprintf("KEYS[%d]", key+1), fmt.Sprintf("'%s'", string(args[key+2])), 1)
			keysArray[key] = args[key+2]
		}
	}

	// replace ARGVs for lua
	argsArray := make([]string, len(args[keysLen:]))
	for key := range argsArray {
		luaScript = strings.Replace(luaScript, fmt.Sprintf("ARGV[%d]", key+1), fmt.Sprintf("'%s'", string(args[key+keysLen])), 1)
		argsArray[key] = args[key+keysLen]
	}
	// TODO make transactional https://redis.io/commands/eval#atomicity-of-scripts
	if cli == nil {
		cli = NewClientContext(ctx)
	}
	L := lua.NewState()
	defer L.Close()
	registerRedisType(L)
	err = L.DoString(luaScript)
	if err != nil {
		resp.ReplyError(ctx.Out, err.Error())
		return
	}
	//if eval result empty
	if L.GetTop() == 0 {
		resp.ReplyNullBulkString(ctx.Out)
		return
	}

	// filter data from lua to redis command without first arg
	GetLuaTop := L.GetTop()
	if GetLuaTop > 1 {
		resp.ReplyArray(ctx.Out, GetLuaTop)
	}
	for i := 1; i <= GetLuaTop; i++ {
		lv := L.Get(i)
		parseLuaValue(ctx, lv)
	}
	return
}
