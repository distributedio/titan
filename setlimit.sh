#!/bin/bash
host=tkv7369.test.zhuaninc.com
port=7369
rediscli="./redis-cli"
token="sys_ratelimit-1574130304-1-36c153b109ebca80b43769"
usage_exit()
{
	echo "usage:"
	echo "$0 set qps=(1|0) cmd=<cmd> namespace=<namespace> limit=<num>[k/K/m/M] burst=<num>"
	echo "or"
	echo "$0 del qps=(1|0) cmd=<cmd> namespace=<namespace>"
	echo "<namespace>: all means matching all namespaces"
	exit 1
}

if [ $# -lt 1 ]; then
	usage_exit
fi
op=$1
if [ "$op" != "set" -a "$op" != "del" ]; then
	usage_exit
fi

limitname=
cmd=
namespace=
limit=
burst=

for arg in $*
do
	if [ "$arg" = "$op" ]; then
		continue
	fi

	optname=`echo $arg | sed 's/=.*//'`
	optvalue=`echo $arg | sed 's/.*=//'`
	if [ -z $optname -o -z $optvalue ]; then
		usage_exit
	else 
		if [ "$optname" = "qps" ]; then
			if [ $optvalue = 1 ]; then
				limitname=qps
			elif [ $optvalue = 0 ]; then	
				limitname=rate
			else
				usage_exit
			fi
		elif [ "$optname" = "cmd" ]; then
			cmd=$optvalue
		elif [ "$optname" = "namespace" ]; then
			namespace=$optvalue
		elif [ "$optname" = "limit" ]; then
			limit=$optvalue
		elif [ "$optname" = "burst" ]; then
			burst=$optvalue
		else
			usage_exit
		fi
	fi
done


if [ -z "$limitname" -o -z "$cmd" -o -z "$namespace" ]; then
	usage_exit
else
	key=
	if [ "$namespace" = "all" ]; then
		key="${limitname}:*@${cmd}"
	else
		key="${limitname}:${namespace}@${cmd}"
	fi	

	if [ "$op" = "set" ]; then
		if [ -z "$limit" -o -z "$burst" ]; then
			usage_exit
		else
			$rediscli -h $host -p $port -a $token set $key "${limit} ${burst}"
		fi
	elif [ "$op" = "del" ]; then
		$rediscli -h $host -p $port -a $token del $key
	else
		usage_exit
	fi
fi
