package command

import (
    "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

func TestNoExpireChange(t *testing.T){
    key := "string_key"
    args := []string{key, "10"}

    ctx := ContextTest("set", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "OK")
    EqualGet(t, key, "10", nil)

    args = []string{key, "a", "b", "c"}
    ctx1 := ContextTest("sadd", args...)
    Call(ctx1)
    assert.Contains(t, ctxString(ctx1.Out), "WRONGTYPE")

    args = []string{key}
    ctx2 := ContextTest("scard", args...)
    Call(ctx2)
    assert.Contains(t, ctxString(ctx2.Out), "WRONGTYPE")

    args = []string{key, "1", "a"}
    ctx3 := ContextTest("zadd", args...)
    Call(ctx3)
    assert.Contains(t, ctxString(ctx3.Out), "WRONGTYPE")
}
func TestExpireChange(t *testing.T){

    Cfg.Expire.Disable = true
    Cfg.GC.Disable = true
    time.Sleep(time.Second)

    key := "string_key"
    args := []string{key, "10", "ex", "4"}

    ctx := ContextTest("set", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "OK")

    args = []string{key, "a", "b", "c"}
    ctx1 := ContextTest("sadd", args...)
    Call(ctx1)
    assert.Contains(t, ctxString(ctx1.Out), "WRONGTYPE")

    time.Sleep(time.Second * 4)

    args = []string{key, "a", "b", "c"}
    ctx2 := ContextTest("sadd", args...)
    Call(ctx2)
    assert.Contains(t, ctxString(ctx2.Out), "3")

    args = []string{key, "4"}
    ctx = ContextTest("expire", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "1")

    args = []string{key, "1", "a"}
    ctx = ContextTest("zadd", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "WRONGTYPE")

    time.Sleep(time.Second * 4)

    args = []string{key, "1", "a"}
    ctx = ContextTest("zadd", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "1")

    args = []string{key, "4"}
    ctx = ContextTest("expire", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "1")

    args = []string{key, "a", "1"}
    ctx = ContextTest("hset", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "WRONGTYPE")

    time.Sleep(time.Second * 4)

    args = []string{key, "0", "2"}
    ctx = ContextTest("zrange", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "0")

    args = []string{key, "1", "a"}
    ctx = ContextTest("hset", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "1")

    args = []string{key, "4"}
    ctx = ContextTest("expire", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "1")

    args = []string{key, "a"}
    ctx = ContextTest("lpush", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "WRONGTYPE")

    time.Sleep(time.Second * 4)

    args = []string{key, "a"}
    ctx = ContextTest("lpush", args...)
    Call(ctx)
    assert.Contains(t, ctxString(ctx.Out), "1")

    Cfg.Expire.Disable = false
    Cfg.GC.Disable = false
    time.Sleep(time.Second)
}
