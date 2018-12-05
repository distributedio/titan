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

### Enable Multi-tenancy
First, set auth value in the 'server' section of conf/titan.toml

```
auth = "YOUR_SERVER_KEY"
```

Then use ./tools/token/token to generate a client token.

```
cd $GOPATH/src/github.com/meitu/titan/tools/token
go build main.go -o titan-gen-client-token
./titan-gen-client-token -key YOUR_SERVER_KEY -namespace bbs
```

Then you'll get the token for client auth, for example: bbs-1543999615-1-7a50221d92e69d63e1b443

### Run

```
cd $GOPATH/src/github.com/meitu/titan
./titan
```

## Check and test

```
redis-cli -p 7369
```

If Multi-tenancy is enabled, use server-generated token to auth:

```
redis-cli -p 7369 -a bbs-1543999615-1-7a50221d92e69d63e1b443
```

