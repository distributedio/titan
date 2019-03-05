# Titan-Set 设计

## 简介

Redis 集合（Set类型）是一个无序的数据的集合，类似List的一个列表，我们可以对集合类型进行元素的添加、删除或判断元素是否存在等操作。

与List相比，区别主要有以下两点：

* Set不能有重复的数据，如果多次添加相同元素，Set中将仅保留该元素的一份拷贝，更新修改时间即可。
* 和List类型相比，Set类型还有一个非常重要的特性，可以在服务器端完成多个集合之间的聚合计算操作，如：[SUNION](https://redis.io/commands/sunion)、[SINTER](https://redis.io/commands/sinter)和[SDIFF](https://redis.io/commands/sdiff)。由于这些操作均在服务端完成，节省了大量的网络I/O开销。

### 应用场景
集合有取交集、并集、差集等操作，因此可以求共同好友、共同兴趣、分类标签等。

## 设计思路

### Set结构体
	type Set struct {
	    meta   *SetMeta
	    key    []byte
	    exist  bool
	    txn    *Transaction
	}
	
### SetMeta信息
	 type SetMeta struct {
	     Object
	     Len int64
	 }
	 
### 设计要点
	 
* 当前实现仍旧选择维护一个Len,记录set内部的元素个数，并且不使用hash中的slot。[hash-slot实现方式](https://github.com/meitu/titan/pull/13#%E8%83%8C%E6%99%AF)
* set集合的特性意味着在kv存储中不需要value值，因此存储时调用TiKV 的Set接口，只需要将拼接好的member存储在tikv对应的key中即可
* MetaKey的具体格式：{Namespace}:{DBID}:M:{key}
* DataKey的具体格式：{Namespace}:{DBID}:D:{ObjectID}
* member在存储中key的格式为：{Namespace}:{DBID}:D:{ObjectID}:{member}
* value部分存储为：[]byte{0}，因为tikv不允许value为空，占位即可。

	 
## 命令处理
### 常规命令处理
#### SAdd key member [member ...]
* 在key对应的set中添加一个元素。

**实现步骤**

* 对传入的member进行去重
* 调用BatchGetValues批量获取datakey对应的value值，根据value的类型判断member是否已经存在
* 过滤掉已经存在的member，统计新增member的数量
* 更新Meta信息并返回新增member的数量

#### SMembers key

* 返回key中的集合值的所有成员

**实现步骤**

* 使用迭代器寻找拼接好的前缀在存储中的位置
* 返回具有相同前缀的所有元素


#### SCard key
* 返回集合key中元素的数量。

**实现步骤**

* 返回meta信息中Len即可


#### SIsmember key member 
* 判断元素member是否是集合key的成员。

**实现步骤**

* 在存储中寻找拼接好的datakey是否存在
* 如果存在则返回1，否则返回0

#### SPop key

* 随机返回并删除key对应的set中的一个元素。因为数据在TiKV中存储有序，因此直接删除并返回key对应set中的第一个元素即可

**实现步骤**

* 删除key对应set中的第一个元素
* 更新meta信息，返回删除的member
 

#### SRem key member [member ...]
* 移除集合key中指定的一个或多个元素member，如果member不存在会被忽略。

**实现步骤**

* 在存储中寻找拼接好的datakey是否存在
* 如果存在则调用delete进行删除
* 更新meta信息

### SMove key key1 member
* 将key对应的源集合的member移动到key1对应的目标集合中，如果key对应的源集不存在或不包含指定的元素，则不执行任何操作并返回0。否则，元素将从key对应的源集删除并添加到key1对应的目标集。

**实现步骤**

* 首先获取判断member是否属于key对应的源集合,如果源集合内部不存在直接返回
* 接下来判断目标key1对应的集合内是否存在member，如果不存在则将member加入到key1对应的目标集合中，更新meta信息
* 删除源key对应的集合中的member,更新meta信息

### 集合命令处理
对于集合类命令而言，最直观的实现方案就是在计算交并差集时将所有的member读到内存在进行计算，虽然通过map进行去重以及排序可以优化部分计算的性能，但是当set的元素个数特别多时，仍会存在内存吃紧的问题。因为每一个key对应set中存储的member在内存中是有序的，因此可以借鉴归并的思想完成集合操作，具体实现思路如下：
#### SUion——求给定key对应的set并集
##### 实现步骤
1. 为每一个key对应的集合设定一个指针(iter)，指向当前 set 的第一个member(key),因为member在存储中有序，因此第一个member一定是当前set中最小的
2. 比较各个member大小，分为以下几种情况
	* 如果大小相同证明是相同的元素，则记录这个member作为并集结果的一部分
	* 如果存在大小不相等的情况，将最小的元素所对应的指针向后移动，并记录member为并集结果的一部分
3. 重复步骤2，直到所有集合的member全部seek完成，若果比较到最后只剩下一个集合则将这个集合的剩余元素作为并集结果的一部分并返回并集结果

#### SInter——求给定key对应的set交集
##### 实现步骤

1. 创建set对象的同时读取meta信息，对于不存在的集合当做空集来处理。一旦出现空集，则不用继续计算了，最终的交集就是空集。
2. 如果不存在空集，则为每一个key对应的集合设定一个指针(iter)，指向当前 set 的第一个member(key),因为member在存储中有序，因此第一个member一定是当前set中最小的
3. 比较各个member大小，分为以下几种情况
	* 如果大小相同证明是相同的元素，则记录这个member作为交集结果的一部分，将所有key对应set的指针向后移动一个位置
	* 如果存在大小不相等的情况，证明较小的元素绝对不会出现在其他的member中，将除了最大的member以外的指针(iter)向后移动一个位置
4. 重复第二步直到某一指针超出序列尾，此时证明member数量最少的set已经全部seek完成，此时退出流程，返回交集结果。


#### SDiff key [key ...]——求给定key对应的set与第一个key对应的set的差集
##### 实现步骤


1. 为每一个key对应的集合设定一个指针(iter)，指向当前 set 的第一个member(key),因为member在存储中有序，因此第一个member一定是当前set中最小的
2. 以指定的第一个key为基准key，与其他的key比较member大小，分为以下几种情况
	* 如果和后续member比较之后大小相同证明是相同的元素，则说明该member不在差集范围中，将基准key以及有相同member的key得指针向后移动一个位置
	* 如果不相等，则将指向最小的member对应key的指针向后移动，如果这个key是基准key，则记录指针此时指向的member为差集的一部分
3. 重复步骤2，直到所有set的member全部seek完成，返回差集结果
