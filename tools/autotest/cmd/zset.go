package cmd

import (
    "testing"
    "sort"
    "strconv"

    "github.com/gomodule/redigo/redis"
    "github.com/stretchr/testify/assert"
)

//ExampleList the key command
//memberScores record the key and value of the operation
type ExampleZSet struct {
    memberScores map[string]map[string]float64
    conn         redis.Conn
}

//NewExampleList create list object
func NewExampleZSet(conn redis.Conn) *ExampleZSet {
    return &ExampleZSet{
        conn:         conn,
        memberScores: make(map[string]map[string]float64),
    }
}

func (ez *ExampleZSet) ZAddEqual(t *testing.T, key string, values ...string) {
    msmap, ok := ez.memberScores[key]
    if !ok {
        ez.memberScores[key] = make(map[string]float64)
        msmap = ez.memberScores[key]
    }
    oldLen := len(msmap)

    req := make([]interface{}, 0, len(values))
    req = append(req, key)
    for i := range values {
        req = append(req, values[i])
    }

    if len(values) %2 != 0 {
        reply, err := redis.Int(ez.conn.Do("zadd", req...))
        assert.Equal(t, oldLen, len(msmap))
        assert.Nil(t, reply)
        assert.NotNil(t, err)
        return
    }

    for i := range values {
        if i%2 == 0 {
            fscore, err := strconv.ParseFloat(values[i], 64)
            if err != nil {
                reply, err := redis.Int(ez.conn.Do("zadd", req...))
                assert.Equal(t, oldLen, len(msmap))
                assert.Nil(t, reply)
                assert.NotNil(t, err)
                return
            }
            msmap[values[i+1]] = fscore
        }
    }

    reply, err := redis.Int(ez.conn.Do("zadd", req...))
    assert.Equal(t, len(msmap)-oldLen, reply)
    assert.Nil(t, err)
}

func (ez *ExampleZSet) ZAnyOrderRangeEqual(t *testing.T, key string, start int, stop int, positiveOrder bool, withScore bool) {
    cmd := "zrange"
    if !positiveOrder {
        cmd = "zrevrange"
    }
    msmap, ok := ez.memberScores[key]
    if !ok {
        reply, err := redis.Strings(ez.conn.Do(cmd, key, start, stop))
        assert.Equal(t, msmap, reply)
        assert.Nil(t, err)
        return
    }

    if start >= len(msmap) {
        reply, err := redis.Strings(ez.conn.Do(cmd, key, start, stop))
        assert.Equal(t, []string{}, reply)
        assert.Nil(t, err)
        return
    }

    tmp := getAllOutput(msmap, positiveOrder, withScore)
    var reply []string
    var err error
    if withScore{
        reply, err = redis.Strings(ez.conn.Do(cmd, key, start, stop, "WITHSCORES"))
    } else {
        reply, err = redis.Strings(ez.conn.Do(cmd, key, start, stop))
    }

    if start < 0 {
        if start += len(msmap); start < 0 {
            start = 0
        }
    }
    if stop < 0 {
        if stop += len(msmap); stop < 0 {
            stop = 0
        }
    } else if stop >= len(msmap){
        stop = len(msmap)-1
    }
    if withScore{
        assert.Equal(t, tmp[2*start:2*stop+2], reply)
    }else{
        assert.Equal(t, tmp[start:stop+1], reply)
    }
    assert.Nil(t, err)
}

func (ez *ExampleZSet) ZRangeEqual(t *testing.T, key string, start int, stop int, withScore bool) {
    ez.ZAnyOrderRangeEqual(t, key, start, stop, true, withScore)
}

func (ez *ExampleZSet) ZRevRangeEqual(t *testing.T, key string, start int, stop int, withScore bool) {
    ez.ZAnyOrderRangeEqual(t, key, start, stop, false, withScore)
}

func getAllOutput(msmap map[string]float64, positiveOrder bool, withScore bool)([]string){
    scoreMembers := make(map[float64][]string)
    for member, score := range msmap {
        if _, ok := scoreMembers[score]; !ok {
            scoreMembers[score] = make([]string, 0, 1)
        }
        scoreMembers[score] = append(scoreMembers[score], member)

    }
    scores := make([]float64, 0)
    for score, _ := range scoreMembers {
        scores = append(scores, score)
    }
    if positiveOrder {
        sort.Float64s(scores)
    } else {
        sort.Sort(sort.Reverse(sort.Float64Slice(scores)))
    }

    fullOutput := make([]string, 0, 2*len(msmap))
    for _, score := range scores {
        members := scoreMembers[score]
        if positiveOrder {
            sort.Strings(members)
        } else {
            sort.Sort(sort.Reverse(sort.StringSlice(members)))
        }
        for _, member := range members {
            fullOutput = append(fullOutput, member)
            if withScore{
                val := strconv.FormatFloat(score, 'f', -1, 64)
                fullOutput = append(fullOutput, val)
            }
        }
    }
    return fullOutput
}