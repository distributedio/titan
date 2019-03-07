# Proposal for Titan-Set

## Introduction

The *Set* type in Redis is a collection of unordered data.  Like an instance of *List* type, operations like adding elements, deleting elements or checking if some element is existed and so on can be easily done with a *Set* instance.

Furthermore,  here we list the main difference between the *List* and *Set*.

+ All the elements in one set are different from each other. If an existed element is added for multiple times, then this element will be continuously updated with its duplications. Actually, it is quite easy to manage just by updating the last modification time.
+ Another important property of *Set* is that some aggregation operations within several sets as [SUNION](https://redis.io/commands/sunion), [SINTER](https://redis.io/commands/sinter) and [SDIFF](https://redis.io/commands/sdiff) can be accomplished on the server side which saves the heavy overhead of Network I/O.

### Applications

Functions of calculating intersections, unions and difference sets are designated for those situations in which digging for common friends, common interest, classifying labels and so on are the basic task.

## Design

### Struct *Set*

type Set struct *{*
	    *meta*     **SetMeta*
	    *key*      *[ ]byte*
	    *exist*    *bool*
	    *txn*      *\*Transaction*
	    *}*

### Struct *SetMeta*

type SetMeta struct *{*
	     *Object*
	     *Len*     *int64*
	 *}*

### Key Points

+ The present implementation is still designed to use  *Len* to keep the count of the elements within a set and we do not use the *slot* in hash. *[[Read More] Implementation of hash-slot](https://github.com/meitu/titan/pull/13#%E8%83%8C%E6%99%AF)*
+ ==The characteristics of set means that we don't need to store the *value*  in the *kv-storage*. So for the case of storing data in Tikv which the Set interface will be called, we only need to save the  concatenated member in the *key* of *tikv*.==
+ The format of *MetaKey* is:
  {Namespace}:{DBID}:M:{key}
+ The format of *DataKey* is:
  {Namespace}:{DBID}:D:{ObjectID}
+ The format of *member* which will be stored in the *key* of *tikv*:
  {Namespace}:{DBID}:D:{ObjectID}:{member}
+ The format of *value* in *tikv* (for an empty *value* is not permitted in *tikv* and we use a placeholder instead):
  []byte{0}

## Command processing
### General command processing
#### SAdd key member [member ...]
* Add an element to the set corresponding to the key.


**Implementation steps**

* Deweight the member passed in
* Call BatchGetValues to batch obtain the corresponding value of datakey, and judge whether the member already exists according to the type of value
* Filter out existing members and count the number of new members
* Update Meta information and return the number of new members


#### SMembers key

* Returns key for all members of the collection
**Implementation steps**

* Use an iterator to find the location of the spliced prefix in the store
* returns all elements with the same prefix

#### SCard key
* returns the key corresponding to the number of elements in the collection.。

**Implementation steps**

* return Len in the meta information

#### SIsmember key member
* Determines whether the element member is a member of the set key。

**Implementation steps**

* Look for the existence of a spliced datakey in the storage
* returns 1 if it exists, or 0 if it does not

#### SPop key

* Returns and deletes an element of the set corresponding to the key at random. Since the data is stored in TiKV in an orderly manner, you simply delete and return the key corresponding to the first element in the set

**Implementation steps**

* Delete the key corresponding to the first element in the set
* Update meta information to return deleted member


#### SRem key member [member ...]
* Removes one or more element members specified in the collection key. Member is ignored if it does not exist.

**Implementation steps**

* Look for the existence of a spliced datakey in the storage
* if there is, call delete to delete
* update meta information

### SMove key key1 member
* Moves the member of the source set corresponding to key into the target set corresponding to key1. If the source set corresponding to key does not exist or contains the specified element, no operation is performed and 0 is returned. Otherwise, the element will
* Delete from the source set corresponding to key and add to the target set corresponding to key1.

**Implementation steps**

* First, get whether member belongs to the source set corresponding to key, if there is no direct return inside the source set
* next, determine whether member exists in the set corresponding to target key1. If not, add member to the set corresponding to key1 to update meta information
* Last,delete member in the set corresponding to the source key and update meta information


### Collective command processing


   For the collection class command, the most intuitive implementation scheme is to read all members into the memory for calculation when calculating the intersection difference set. Although the performance of partial calculation can be optimized by deduplication and sorting through the map, the memory is still tight when the number of elements in the set is particularly large. Because each key corresponds to the member stored in set, which is in order in memory, we can refer to the idea of merge to complete the set operation, and the specific implementation idea is as follows：

#### SUion -- find the set union corresponding to the given key
##### Implementation steps

1. Set a pointer (iter) for the set corresponding to each key, pointing to the first member(key) of the current set. Since member is stored in order, the first member must be the smallest in the current set
2. Compare each member size, divided into the following two situations
   
   * if the same size proves to be the same element, record this member as part of the union result
   * if an unequal size exists, move the pointer to the smallest element backward and record member as part of the union result

3. Repeat step 2 until all members of the collection have completed their seek. If only one set is left after the comparison, the remaining elements of the set will be taken as part of the union result and the union result will be returned

#### SInter -- find the set intersection corresponding to the given key
##### Implementation steps

1. Meta information is read while set object is created, and nonexistent collection is treated as empty set. Once there is an empty set, there is no need to continue the calculation, and the final intersection is the empty set.
2. If no empty set exists, a pointer (iter) is set to the corresponding set of each key, pointing to the first member(key) of the current set. Since member is stored in an orderly way, the first member must be the smallest in the current set
3. Compare the size of each member and divide it into the following situations
	* if the same size proves to be the same element, record this member as part of the result of the intersection, and move the pointer of all keys corresponding to the set back one position
	* in case of unequal size, it is proved that the smaller element will never appear in other members, and the pointer (iter) except the largest member will be moved back one place
4. Repeat the second step until a pointer exceeds the end of the sequence. At this time, it is proved that the set with the lowest number of members has completed all seek.


#### SDiff key [key]... -- find the difference between the set corresponding to the given key and the set corresponding to the first key
##### Implementation steps


1. Set a pointer (iter) to the set corresponding to each key, pointing to the first member(key) of the current set. Since member is stored in an orderly way, the first member must be the smallest > in the current set
2. Take the specified first key as the benchmark key, compare the member size with other keys, and divide into the following situations
* if the member is of the same size after comparison with the following member, it proves to be the same element, indicating that the member is not in the range of difference set. Move the pointer of the benchmark key and the key with the same member back one > position
* if not, the pointer to the smallest member corresponding to key is moved backward; if the key is the reference key, the record pointer points to member as part of the difference set
3. Repeat step 2 until all members of set complete seek and return the result of difference set



