#!/bin/bash

usage_exit()
{
	echo "usage:"
	echo "<hostportconfigpath> set qps=(1|0) cmd=<cmd> namespace=<namespace> limit=<num>[k/K/m/M] burst=<num>"
	echo "or"
	echo "<hostportconfigpath> del qps=(1|0) cmd=<cmd> namespace=<namespace>"
	echo "<namespace>: all means matching all namespaces"
	exit 1
}
if [ $# -lt 5 ]; then
	usage_exit
fi

configpath=$1
host=`grep host= $configpath|sed 's/host=//'`
port=`grep port= $configpath|sed 's/port=//'`
rediscli=`grep rediscli= $configpath|sed 's/rediscli=//'`
token=`grep token= $configpath|sed 's/token=//'`

op=$2
if [ "$op" != "set" -a "$op" != "del" ]; then
	usage_exit
fi

limitname=
cmd=
namespace=
limit=
burst=

i=0
for arg in $*
do
	i=`expr $i + 1`
	if [ $i -le 2 ]; then
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
