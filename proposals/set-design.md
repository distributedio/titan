# Titan-Set 设计

## 简介

Redis 集合（Set类型）是一个无序的String类型数据的集合，类似List的一个列表，我们可以对集合类型进行元素的添加、删除或判断元素是否存在等操作。

与List相比，区别主要有以下两点：

* Set不能有重复的数据，如果多次添加相同元素，Set中将仅保留该元素的一份拷贝。
* 和List类型相比，Set类型还有一个非常重要的特性，可以在服务器端完成多个集合之间的聚合计算操作，如：SUNION、SUNIONSTORE和SDIFFSTORE。由于这些操作均在服务端完成，因此效率极高，而且也节省了大量的网络I/O开销。

### 应用场景
集合有取交集、并集、差集等操作，因此可以求共同好友、共同兴趣、分类标签等。

#### redis的Set实现方式

在reids中，使用intset实现集合(set)这种对外的数据结构。与Redis对外暴露的其它数据结构类似，set的底层实现，随着元素类型是否是整型以及添加的元素的数目多少，而有所变化。概括来讲，当set中添加的元素都是整型且元素数目较少时，set使用intset作为底层数据结构，否则，set使用dict作为底层数据结构。

* intset数据结构简介：[intset简介](https://juejin.im/post/58350d1a67f3560065e74bde)
* dict数据结构简介：[dict简介](https://mp.weixin.qq.com/s?__biz=MzA4NTg1MjM0Mg==&mid=2657261203&idx=1&sn=f7ff61ce42e29b874a8026683875bbb1&scene=21#wechat_redirect)


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
	 
* 当前实现仍旧选择维护一个Len,并且不使用slot，其他实现方式按照hash相似的方式进行
* set不需要value值，因此存储时调用kvstore.set()接口，只需要将member存储在tikv对应的key中即可value部分存储为：[]byte{0}，因为tikv不与许value为空，占位即可。
	 
## 命令处理
### 常规命令处理
#### SAdd key member [member ...]
* 在key对应的set中添加一个元素。

**实现步骤**

* member去重复
* BatchGetValues批量获取datakey值。
* 过滤掉已经存在的member，统计add的数量
* 更新Meta信息并返回add的数量

#### SMembers key

* 返回存储在key中的集合值的所有成员

**实现步骤**

* 拼接datakey
* 根据前缀寻找到对应的key在存储中的位置
* 返回具有相同前缀的所有元素




#### SCard key
* 返回集合key中元素的数量。

**实现步骤**
* 返回meta信息中Len即可


#### SIsmember key member 
* 判断元素member是否是集合key的成员。

**实现步骤**

* 拼接datakey
* 根据前缀寻找到对应的key在存储中的位置
* 遍历所有相同前缀的key，与拼接member后相同则返回1，否则返回0



#### SPop key
* 随机返回并删除key对应的set中的一个元素。

**实现步骤**


#### SRem key member [member ...]
* 移除集合key中指定的一个或多个元素member，如果member不存在会被忽略。

**实现步骤**

* 拼接datakey
* 根据前缀寻找到对应的key在存储中的位置
* 遍历寻找到要删除的member，调用delete删除
* 更新meta信息

### SMove key key1 member
* 将key对应的源集合的member移动到key1对应的目标集合中，如果key对应的源集不存在或不包含指定的元素，则不执行任何操作并返回0。否则，元素将从key对应的源集删除并添加到key1对应的目标集。

**实现步骤**

* 首先获取判断member是否属于key对应的源集合,如果源集合内部不存在直接返回
* 接下来判断目标key1对应的集合内是否存在member，如果不存在则将member加入到key1对应的目标集合中，更新meta信息
* 删除源key对应的集合中的member,更新meta信息

### 集合命令处理


#### SInter——求给定key对应的set交集
##### 实现思路
1. 读取meta信息，对于不存在的集合当做空集来处理。一旦出现空集，则不用继续计算了，最终的交集就是空集。
2. 如果不存在空集，取出第一个集合的全部member为基准集合。
3. 从第二个集合开始遍历后续每个集合的member，将两个集合中都存在的元素作为下一次操作的基准，直至操作全部完成。

#### SDiff key [key ...]——求给定key对应的set与第一个key对应的set的差集
##### 实现思路

1. 将第一个集合的所有元素都加入到一个中间集合中。
2. 遍历后面所有的集合，对于碰到的每一个元素，从中间集合中删掉它。
3. 最后中间集合剩下的元素就构成了差集。

#### SUion——求给定key对应的set并集
##### 实现思路
1. 将所有集合的全部member加入到一个集合中
2. 最后对集合进行去重














