# Titan-Set 设计

## 简介

Redis 集合（Set类型）是一个无序的数据的集合，类似List的一个列表，我们可以对集合类型进行元素的添加、删除或判断元素是否存在等操作。

与List相比，区别主要有以下两点：

* Set不能有重复的数据，如果多次添加相同元素，Set中将仅保留该元素的一份拷贝，重设更新修改时间即可。
* 和List类型相比，Set类型还有一个非常重要的特性，可以在服务器端完成多个集合之间的聚合计算操作，如：SUNION、SUNIONSTORE和SDIFFSTORE。由于这些操作均在服务端完成，节省了大量的网络I/O开销。

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
	 
* 当前实现仍旧选择维护一个Len,记录set内部的元素个数，并且不使用hash中的slot。[hash-slot实现方式]()
* set集合的特性意味着在kv存储中不需要value值，因此存储时调用kvstore.set()接口，只需要将拼接好的member存储在tikv对应的key中即可
* MetaKey的具体格式：{Namespace}:{DBID}:M:{key}
* DataKey的具体格式：{Namespace}:{DBID}:D:{key}
* member在存储中key的格式为：{Namespace}:{DBID}:D:{ObjectID}:{member}
* value部分存储为：[]byte{0}，因为tikv不允许value为空，占位即可。

	 
## 命令处理
### 常规命令处理
#### SAdd key member [member ...]
* 在key对应的set中添加一个元素。

**实现步骤**

* 对传入的member进行去重
* 调用BatchGetValues批量获取datakey对应的value值，根据value的类型判断member是否已经存在
* 过滤掉已经存在的member，统计add member的数量
* 更新Meta信息并返回add member的数量

#### SMembers key

* 返回存储在key中的集合值的所有成员

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

* 随机返回并删除key对应的set中的一个元素。

**实现步骤**

* 创建随机数，随机数小于set的长度
* 根据参数确定删除的member个数
* 使用迭代器寻找拼接好的前缀在存储中的位置，向后迭代随机数次，然后删除key
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

#### SUion——求给定key对应的set并集
##### 实现思路
1. 遍历参数key,取出每个key对应集合的member并依次放入到一个map中，key对应的member作为map的键值，
2. 直接遍历map的键值得到结果

#### SInter——求给定key对应的set交集
##### 实现思路

1. 使用BatchGetValues批量读取meta信息，对于不存在的集合当做空集来处理。一旦出现空集，则不用继续计算了，最终的交集就是空集。
2. 如果不存在空集，则获取每个key对member个数，选取最少个数的集合为基准集合，这个排序有利于后面计算的时候从最小的集合开始，需要处理的元素个数较少。
3. 遍历基准集合，对于它的每一个元素，依次在后面的所有集合中进行查找。只有在所有集合中都能找到的元素，才加入到最后的结果集合中。

#### SDiff key [key ...]——求给定key对应的set与第一个key对应的set的差集
##### 实现思路

1. 将第一个集合的所有元素都加入到一个中间集合中。
2. 遍历后面所有的集合，对于碰到的每一个元素，从中间集合中删掉它。
3. 最后中间集合剩下的元素就构成了差集。
