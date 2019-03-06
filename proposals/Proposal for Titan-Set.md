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



