package server

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const tlsCert = `-----BEGIN CERTIFICATE-----
MIIEIDCCAgigAwIBAgIQNAgWekQnAMTf7Vc5SVAxTzANBgkqhkiG9w0BAQsFADAR
MQ8wDQYDVQQDEwZyb290Q0EwHhcNMTkwNjA4MjE0NjIxWhcNMjAxMjA4MjE0NTE5
WjAUMRIwEAYDVQQDEwlsb2NhbGhvc3QwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAw
ggEKAoIBAQDMV1/foVWSm74E0xD9hy94EpDoFZAXRg5Y20Em1azh6FQrl2Il14KA
zT4IfF3tFt8AGDn5rMo37bX9BUxgHAUA6cJZ63YRkkdHPQoa8plG88dmRsp3cC83
quO+8BZgWlWno4KKLDA8BKc2VETlmszvQB/JM0FnuwN4z0ODu+yJ6CFfhWX6QTgM
oe9MMxCJxL/xdqNoKi/QikggTlRajuSz6KpYWvB6BQpcxbAZUbCZ8AK70RzROLtb
Z0nWVH39ca7wWF+RHTyxlGydTEMQXMMVJrDOm3h60qyXbAzuRU5C3YEC2A2FZhU1
CQz3iCT3rDGjpbmiO/uwO9cvRR2Ng7nbAgMBAAGjcTBvMA4GA1UdDwEB/wQEAwID
uDAdBgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwHQYDVR0OBBYEFLxmWuLc
RnFxh1D70l70qddmVz33MB8GA1UdIwQYMBaAFAJNm9u0J4dfLnJSkARgp5Ix/v2n
MA0GCSqGSIb3DQEBCwUAA4ICAQDGUlyn0RkYFk+IFEpDrlptjg+x+WNjNMMlNyz0
BtG6WYj6bz6xnKfUeui3ei+0GXfj2Q+fbgw1Cs+1rNVRl8e3wSybT5ZmZWK2jOm/
n41TdtwUtBJoktklmd8/EyXz1tHdIv0F7YVP0glXWmJe60K1N7M4jc/hKDF6u9fl
hhJX20FtjSLAM9wWAZV+2YFnIGviqBg7YXiTxpqO4usdFO3RnnwwLslc0v1e2ME6
RTI1/pXmIOles77S+a5UQQr02nyXZ6nnSQXJlSbAFB6giLlVCyWXwHR3jCqX9to5
mj8WFjSreWVLxATRXiPqmr69b2W6sVp8TcswwlqOxFRBNOceRuabyMHoaA+N4kZQ
+aI73Ad1Xlz0H6L+DE4ATo8FWryNmL+9OgNxgP1SW/dM4FRjn2nTXyg+t+0lgY3w
JviOJj5hc5Nr8FmoTQXjq2PbMn4qjPYeXT4geyhyCipY7WDAqUUBHYSohy4IzhDN
DZ8dAV1uSXjUwfzwrqvTJesTNYmaHosdM+GmSLWTzFuHiLXMZ1HYRiZvNZ/tkh6F
xQ0a1lTbVveA7MlvLqOpC9lGP2/UJgbsN/Cy4xbKmenFeh+u6Ez0etqui8z17Tmf
9/dmx6E1BSgOjcwJ7NYCCtPI/SpkQ8WAJRwoUgx/ae9JCD6SBQQZalA0auZNOWXq
3YrAdg==
-----END CERTIFICATE-----`

const tlsKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAzFdf36FVkpu+BNMQ/YcveBKQ6BWQF0YOWNtBJtWs4ehUK5di
JdeCgM0+CHxd7RbfABg5+azKN+21/QVMYBwFAOnCWet2EZJHRz0KGvKZRvPHZkbK
d3AvN6rjvvAWYFpVp6OCiiwwPASnNlRE5ZrM70AfyTNBZ7sDeM9Dg7vsieghX4Vl
+kE4DKHvTDMQicS/8XajaCov0IpIIE5UWo7ks+iqWFrwegUKXMWwGVGwmfACu9Ec
0Ti7W2dJ1lR9/XGu8FhfkR08sZRsnUxDEFzDFSawzpt4etKsl2wM7kVOQt2BAtgN
hWYVNQkM94gk96wxo6W5ojv7sDvXL0UdjYO52wIDAQABAoIBAQCew1g7IVeiRB08
FF2EDc+k5A/wMii03HpzMU8KhEQBdYhIIiNgsXO07UJAR5iWiAmVQj1xLn4jPC8E
umQf3EVK81RMlvQyLMvynotGaq0KgoevgFr4t5IIF19Bz7oi/KzGRfU7s596Ukc0
n/6zwjVtwg2wPoGXvaax659SL+VVM0Gg4EQsBEb0wA+MFHeLf3YBoiPSeK95iG8M
Hnr5pzWJEX5smMDLx/HEow503s7H1QSKYeJuKNXnypkFcN1xqFVOpgIW9n0DxcMl
YyJT749yY05bW64RrLKy207CrY/fUjFfFIptya47fgj+WQUp06VJ4kUcYEsWzcbm
OUikHyZBAoGBAOF6sVJ4HcBxjoyuHhVqqkv1WVn4xvCDLDS6WDAT6k3DphC9QO7X
DAvtwBj2OZSSVr3yhWJjVM+EtFMJWPo4otZUK8Phj1Tj1rlpdH1OLHUxByfs4dwY
Crot4pSJYfcQy7jPiC3lKRJud2OTyH7O6wooPaOPXe7C4Q3eRunSc8i7AoGBAOgA
NY+c8YNgDoWk38kuvoA5/keGzbyrTvri0Gd0N5XdwirtkDK2mgZhHxW3l8WaJiZo
V3GHaLb4aqc0I6Kvk9zSwdqCvY/o9Nz4k4EWGEFfZqgDp/T0hvAyg9YWoH2v+uxR
OgQawAo0oG4oW4wPzbPYQiXltmkeHjC4CFagDNFhAoGAJReG5hcmZcsIdTILduB2
JUq2KSvYpiYd9oqVCUutZp+ByQ0pCmFL9QZmbHTM4hj0tgiYUqgegojFFUfbYEZC
21k7XdzUNFXKs/OaGybp/1lSYQoB2bAGy7vSoza6a+dSbBOPxmUFTafocfQUrm+h
kKkwAqEKBcX/OcXQCpT5QRMCgYB1Io3oaaQi4Z/TaDA5Amnake1Jrc04ggHJeDUi
1rGt8B410GYqxLk1mVm5fE2bzj2OzMXBo02CfCBVNWT8oct1BdAshDAzdboTy0mm
NkKe1w0crWPisIdkxQx9TkVP0EdPg59YLS1iubl6hNPb/qqsL/cN7VJQ9ozlqjVD
j2GJYQKBgQCuy7rDekJDWGj86UF3TyEPa00BXCtwqWX7O5vgAamMU2QLP1dCKLGy
DeywM6K4S9YeP4E9wj0f6Ux5pgqZEZjoaMzz5LM4CH/HDvGtCEzjutWk8lpOeqKn
7bEcpNch2r0zAbYL4fnE27dOdbcEjOf5+SPwnkeEx+yQEXCBnIr9zw==
-----END RSA PRIVATE KEY-----`

func TestTLSConfig(t *testing.T) {
	// setup test files
	cert, _ := ioutil.TempFile("", t.Name()+"_cert")
	_, _ = cert.WriteString(tlsCert)
	defer os.Remove(cert.Name())

	key, _ := ioutil.TempFile("", t.Name()+"_key")
	_, _ = key.WriteString(tlsKey)
	defer os.Remove(key.Name())

	broken, _ := ioutil.TempFile("", t.Name()+"_broken")
	_, _ = broken.WriteString("not a valid cert/key")
	defer os.Remove(broken.Name())

	// not found file #1
	_, err := TLSConfig("/not/found", key.Name())
	assert.Error(t, err)

	// not found file #2
	_, err = TLSConfig(cert.Name(), "/not/found")
	assert.Error(t, err)

	// broken file #1
	_, err = TLSConfig(broken.Name(), key.Name())
	assert.Error(t, err)

	// broken file #2
	_, err = TLSConfig(cert.Name(), broken.Name())
	assert.Error(t, err)

	// success
	opts, err := TLSConfig(cert.Name(), key.Name())
	assert.NoError(t, err)
	assert.NotNil(t, opts)
	assert.Len(t, opts.Certificates, 1)
}
