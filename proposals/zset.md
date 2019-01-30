 # 存储结构设计
 
* meta key


设计与list仍然基本相同，不再赘述，只是value部分除了包含ID、时间等的object外，只存一个len

* 每个member会有两个key

memberKey：存member->score映射的，{db.ns}:{db.ID}:D:{obj.id}:{member}  ->  {score}

scoreKey：可以根据score排序的， {db.ns}:{db.ID}:S:{obj.id}:{score}:{member}   -> {nil}

中间前缀与上一个不同，使用了S而不是D，这是为了zrange等涉及score排序的遍历操作时，能直接seek到这些元素，而不会前一种key包含进来

# 命令处理
 
	
* zadd

查询metaKey得到objId，对每一个member

    查询memberKey，如果存在：查询scoreKey，score相同不作处理，不相同写入新scoreKey，删除老scoreKey，命令返回值不增加

    如果不存在，写入memberKey和scoreKey，命令返回值+1，

如果返回值>0，len+返回值覆盖metaKey的value

* zcard

查询metaKey，返回member的数量
	
* zrem

查询metaKey得到objId，对每一个member
    
    查询memberKey是否存在，存在则删除memberKey和scoreKey，命令返回值+1

    如果不存在，命令返回值不增加

如果返回值>0，len-返回值覆盖metaKey的value，

如果返回值等于len，删除metaKey,写入GC
	
* zrange

查询metaKey得到objId，seek到{db.ns}:{db.ID}:D:{obj.id}:为前缀的，对排序号start和end之间的返回member及其score

* zverrange

查询metaKey得到objId，反向seek到{db.ns}:{db.ID}:D:{obj.id}:为前缀的，对排序号start和end之间的返回member及其score
	
* zrank

查询metaKey得到objId，在scoreKey中seek，即{db.ns}:{db.ID}:S:{obj.id}为前缀的，循环判断key的后一部分为:XXX:{member}的，返回其排名数
	
* zcore

查询metaKey得到objId，拼凑memberKey，返回score值
	
* del

无需修改，只是删除metaKey，生成GC key，GC处理流程中修改增加对zset删除的支持
	
* expire

同上
