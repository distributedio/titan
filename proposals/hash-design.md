# Hashes 设计

## 背景
Hash是Redis最常用和重要的数据结构，其Meta信息维护了Hash对象的成员个数和更新时间，在并发写入（比如HSET）的场景下，可能造成大量事务冲突，严重影响Hash结构的并发性能。如果选择不维护上述两个Meta信息，则在执行HLEN时需要遍历所有相关数据，代价较大。本设计通过引入MetaSlot的概念，将Meta信息分拆到多个Slot中，从而减少并发更新Meta信息的事务冲突，在获取长度和减少并发冲突之间，做出最合适的折中和平衡。

## Meta信息
* Len        标识Hash中成员的个数
* MetaSlot   标识Meta信息Slot个数（注：这里没有直接叫Slot，是因为容易理解为哈希表的slot）

### MetaSlotKey
为了减少并发修改MetaKey造成的事务冲突，我们引入MetaSlotKey，将写Meta信息的请求，均衡分散到多个MetaSlotKey上。

MetaSlotKey 信息
* Len
* UpdatedAt

MetaSlotKey 格式
```
{namespace}:{dbid}:MS:{objectid}
```

新增tag:MS作为MetaSlotKey的标识，既可以避免遍历Meta信息时被MetaSlotKey干扰，也可以将Meta信息与MetaSlot信息尽可能放到同一个Region。

### 均衡算法
随机。每次写入，随机0~MetaSlot-1之间的数字，选择对应的MetaSlotKey，更新Meta信息。

## 命令处理

### HSET、HSETNX、HMSET

1. 获取Meta信息，判断Slot个数，如果为0，则更新Meta信息时，直接写入MetaKey
2. 如果Slot个数大于零，在更新Meta信息时，选择对应的MetaSlotKey写入。

### HLEN

1. 获取Meta信息，判断Slot个数，如果为0，则返回Meta信息中的长度。
2. 如果Slot个数大于零，发起一次Seek请求，获取所有MetaSlotKey，返回所有MetaSlotKey中所有长度之和

### HDEL

1. 先获取HASH长度，方法跟HLEN相同
2. 判断HDEL删除的是最后一个Key，则销毁整个HASH对象

### DEBUG OBJECT

1. 获取Meta信息，判断Slot个数，如果为0，返回Meta信息
2. 如果Slot个数大于0，发起一次Seek请求，获取所有MetaSlotKey，将MetaSlotKey中Len加和，作为返回的Len，选择最大的UpdatedAt，作为返回的UpdatedAt

### DEL

1. 获取Meta信息，判断对象类型，如果是HASH，则继续
2. 直接删除Meta信息，并将DataKey前缀交给GC
3. 判断Slot个数，如果大于0，则将MetaSlotKey前缀交给GC

## 设置MetaSlot

### 配置文件
默认为0，可通过配置文件修改，对整个集群生效。修改后，只对新的Key，或原有MetaSlot=0的Key生效。

### 扩展命令

新增扩展命令
```
hmslot key count
```

成功返回SimpleString OK，否则返回Error

### MetaSlot 调整策略

#### MetaSlot增大
修改Meta信息中的MetaSlot为新的值即可。

#### MetaSlot减小
先获取旧的MetaSlotKey，并得到Len和UpdatedAt信息，跟新的MetaSlot一起写入Meta信息。
