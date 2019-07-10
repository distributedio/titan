array set ::toTestCase {
     #string  test
    "SET and GET an item"  1
    "SET and GET an empty item" 1
    "Very big payload in GET/SET" 1
    "Very big payload random access" 1
    "SET 10000 numeric keys and access all them in reverse order" 1
    "SETNX target key missing" 1
    "SETNX target key exists" 1
    "SETNX against not-expired volatile key" 1
    "SETNX against expired volatile key" 1
    "MGET" 1
    "MGET against non existing key" 1
    "MGET against non-string key" 1
    "MSET base case" 1
    "MSET wrong number of args" 1
    "MSETNX with already existent key" 1
    "MSETNX with not existing keys" 1
    "STRLEN against non-existing key" 1
    "STRLEN against integer-encoded value" 1
    "STRLEN against plain string" 1
    "SETRANGE against non-existing key" 1
    "SETRANGE against string-encoded key" 1
    "SETRANGE against key with wrong type" 1
    "SETRANGE with out of range offset" 1
    "GETRANGE against non-existing key" 1
    "GETRANGE against string value" 1
    "GETRANGE fuzzing" 1
    "Extended SET can detect syntax errors" 1
    "Extended SET NX option" 1
    "Extended SET XX option" 1
    "Extended SET EX option" 1
    "Extended SET PX option" 1
    "Extended SET using multiple options at once" 1
    "GETRANGE with huge ranges, Github issue #1844" 1

    #set  test
    "SADD, SCARD, SISMEMBER, SMEMBERS basics - regular set" 1
    "SADD against non set" 1
    "SADD an integer larger than 64 bits" 1
    "Variadic SADD" 1
    "SREM basics - regular set" 1
    "SREM with multiple arguments" 1
    "SREM variadic version with more args needed to destroy the key" 1
    "Generated sets must be encoded as hashtable" 1
    "SINTER with two sets - hashtable" 1
    "SINTER against three sets - hashtable" 1
    "SUNION with non existing keys - hashtable" 1
    "SDIFF with two sets - hashtable" 1
    "SDIFF with three sets - hashtable" 1
    "SDIFF with first set empty" 1
    "SDIFF with same set two times" 1
    "SDIFF fuzzing" 1
    "SINTER against non-set should throw error" 1
    "SUNION against non-set should throw error" 1
    "SINTER should handle non existing key as empty" 1
    "SPOP basics - hashtable" 1
    "SPOP with <count>=1 - hashtable" 1
    "SPOP with <count>" 1
    "SPOP new implementation: code path #1" 1
    "SPOP new implementation: code path #2" 1
    "SPOP new implementation: code path #3" 1
     #list  test
    "LPUSH, RPUSH, LLENGTH, LINDEX, LPOP - regular list" 1
    "R/LPOP against empty list" 1
    "Variadic RPUSH/LPUSH" 1
    "DEL a list" 1
    "LPUSHX, RPUSHX - generic" 1
    "LINSERT raise error on bad syntax" 1
    "LLEN against non-list value error" 1
    "LLEN against non existing key" 1
    "LINDEX against non-list value error" 1
    "LINDEX against non existing key" 1
    "LPUSH against non-list value error" 1
    "RPUSH against non-list value error" 1
    "LRANGE basics - linkedlist" 1
    "LRANGE inverted indexes - linkedlist" 1
    "LRANGE out of range indexes including the full list - linkedlist" 1
    "LRANGE out of range negative end index - linkedlist" 1
    "LRANGE against non existing key" 1
    "LSET - linkedlist" 1
    "LSET out of range index - linkedlist" 1
    "LSET against non existing key" 1
    "LSET against non list value" 1


    #list3  test
    "Explicit regression for a list bug" 1
    "Regression for quicklist #3343 bug" 1

    #zset  test
    "Check encoding - hashtable" 1
    "ZSET basic ZADD and score update - hashtable" 1
    "ZSET element can't be set to NaN with ZADD - hashtable" 1
    "ZSET element can't be set to NaN with ZINCRBY" 1
    "ZADD with options syntax error with incomplete pair"  1
    "ZADD XX returns the number of elements actually added" 1
    "ZADD - Variadic version base case" 1
    "ZADD - Return value is the number of actually added items" 1
    "ZADD - Variadic version does not add nothing on single parsing err" 1
    "ZADD - Variadic version will raise error on missing arg" 1
    "ZCARD basics - hashtable" 1
    "ZREM removes key after last element is removed" 1
    "ZREM variadic version" 1
    "ZREM variadic version -- remove elements after key deletion" 1
    "ZRANGE basics - hashtable" 1
    "ZREVRANGE basics - hashtable" 1
    "ZSET commands don't accept the empty strings as valid score" 1
    "ZSCORE - hashtable" 1
    "ZSET sorting stresser - hashtable" 1
    "ZSETs skiplist implementation backlink consistency test - hashtable" 1

    #hash test
    "HSET/HLEN - Small hash creation" 1
    "HSET/HLEN - Big hash creation" 1
    "Is the big hash encoded with an hash table?" 1
    "HGET against the small hash" 1
    "HGET against the big hash" 1
    "HGET against non existing key" 1
    "HSET in update and insert mode" 1
    "HSETNX target key missing - small hash" 1
    "HSETNX target key exists - small hash" 1
    "HSETNX target key missing - big hash" 1
    "HSETNX target key exists - big hash" 1
    "HMSET wrong number of args" 1


    "HMGET against non existing key and fields"  1
    "HMGET against wrong type" 1
    "HMGET - small hash" 1
    "HMGET - big hash" 1
    "HKEYS - small hash" 1
    "HKEYS - big hash" 1
    "HVALS - small hash" 1
    "HVALS - big hash" 1
    "HGETALL - small hash" 1
    "HGETALL - big hash" 1
    "HDEL and return value" 1
    "HDEL - more than a single value" 1
    "HDEL - hash becomes empty before deleting all specified fields" 1
    "HEXISTS" 1
    "Is a ziplist encoded Hash promoted on big payload?" 1
    "HINCRBY against non existing database key" 1
    "HINCRBY against non existing hash key" 1
    "HINCRBY against hash key created by hincrby itself" 1
    "HINCRBY against hash key originally set with HSET" 1
    "HINCRBY over 32bit value" 1
    "HINCRBY over 32bit value with over 32bit increment" 1
    "HINCRBY fails against hash value with spaces (left)" 1
    "HINCRBY fails against hash value with spaces (right)" 1
    "HINCRBY can detect overflows" 1
    "HINCRBYFLOAT against non existing database key" 1
    "HINCRBYFLOAT against non existing hash key" 1
    "HINCRBYFLOAT against hash key created by hincrby itself" 1
    "HINCRBYFLOAT against hash key originally set with HSET" 1
    "HINCRBYFLOAT over 32bit value" 1
    "HINCRBYFLOAT over 32bit value with over 32bit increment" 1
    "HINCRBYFLOAT fails against hash value with spaces (left)" 1
    "HINCRBYFLOAT fails against hash value with spaces (right)" 1
    "HSTRLEN against the small hash" 1
    "HSTRLEN against the big hash" 1
    "HSTRLEN against non existing field" 1






    "Test HINCRBYFLOAT for correct float representation (issue #2846)" 1

    #expire test
    "EXPIRE - set timeouts multiple times" 1
    "EXPIRE - It should be still possible to read 'x'" 1
    "EXPIRE - After 2.1 seconds the key should no longer be here" 1
    "EXPIRE - write on expire should work" 1
    "EXPIREAT - Check for EXPIRE alike behavior"  1
    "SETEX - Set + Expire combo operation. Check for TTL"  1
    "SETEX - Check value" 1
    "SETEX - Overwrite old key" 1
    "SETEX - Wait for the key to expire" 1
    "SETEX - Wrong time parameter" 1
    "EXPIRE pricision is now the millisecond" 1
    "PEXPIRE/PSETEX/PEXPIREAT can set sub-second expires" 1
    "TTL returns time to live in seconds" 1
    "PTTL returns time to live in milliseconds" 1
    "TTL / PTTL return -1 if key has no expire" 1
    "TTL / PTTL return -2 if key does not exit" 1
    "5 keys in, 5 keys out" 1
    "EXPIRE with empty string as TTL should report an error" 1
}