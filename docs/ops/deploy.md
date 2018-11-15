# Deploy titan

## Deploy TiKV first

Titan is a redis protocol layer running on TiKV, so you should deploy tikv first.

Following this reference to deploy tikv: https://pingcap.com/docs/op-guide/ansible-deployment/


## Deploy Titan

### Build the binary

```
go get github.com/meitu/titan
cd $GOPATH/src/github.com/meitu/titan
make 
```

### Edit the configration file

Edit conf/titan.toml and set the pd-addrs

```
pd-addrs="tikv://your-pd-addrs:port"
```

### Run

```
./titan
```

## Check and test

```
redis-cli -p 7369
```
