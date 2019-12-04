#!/bin/bash

usage_exit()
{
	echo "usage: <hostportconfigpath>"
	exit 1
}
if [ $# -lt 1 ]; then
	usage_exit
fi

configpath=$1
host=`grep host= $configpath|sed 's/host=//'`
port=`grep port= $configpath|sed 's/port=//'`
rediscli=`grep rediscli= $configpath|sed 's/rediscli=//'`
token=`grep token= $configpath|sed 's/token=//'`

for prefix in qps: rate: limiter_status:
do
	echo "-------------------------- $prefix data --------------------------"
	$rediscli -h $host -p $port -a $token scan 0 count 256 match ${prefix}*|sed '1d'|grep -v '^$'|while read key
	do 
		value=`$rediscli -h $host -p $port -a $token get $key`
		echo -e "key=${key}\tvalue=$value"
	done
	echo
done
