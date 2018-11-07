# Configo
Configo is a go library to parse toml configuration using struct tags

## Features
* Configuring parser behaviour using struct tags
* Setting default value or as required
* Validating value using regex, range expression or named validator
* Generating toml template with human friendly information based on go struct and tags
* Building conf file generation tools using configo-build

## Toml
[shafreeck/toml](https://github.com/shafreeck/toml) is a modification version of [naoina/toml](https://github.com/naoina/toml),
adding the abililty to parse complex struct tags and with bugs fixed.

## Validation
configo has a builtin validator with regex and range support

```go
> 1 //greater than 1
>=1 //greater than or equal to 1
>1 <10 //greater than 1 and less than 10, space indicates the "and" of rules

(1, ) //range expression, same as >1
(1, 10) // >1 <10
(1,10]  // >1 <=10

/[0-9]/ //regex matching

/[0-9]+/ (1, 10) // the value should satisfy both the regex and range expression

netaddr //named validator, used to validate a network address
numeric  >10 ( ,100)// mix different expressions, 'and' is used to combine all expressions
```

See the [source code](https://github.com/shafreeck/configo/blob/master/rule/named.go#L12) for all valid "named validators"

## Struct tags

`tags` has a key 'cfg' and its value consists of four parts: "name; default value or required; rule; descripion".
All four parts are splited by ";".

For example:
```go
Listen `cfg:"listen; :8804; netaddr; The listen address of server"`
```

It looks like this when being marshaled to toml
```toml
#type:        string
#rules:       netaddr
#description: The listen address of server
#default:     :8804
#listen=":8804"
```

You can see that we have rich information about the option. And the option is commented out too because it has a default value.

## configo-build
Configo comes with a util tool called configo-build to build a configration file generator for you.

You can use the generator to generate your toml file or update it when you changed your source code(the configuration struct).

```sh
configo-build ./conf.Config
#build a conf generator, the format of arg is "package.struct" package can be
#absolute or relative(golang takes it as an absolute package if it is without
#the prefix "./" or "../").

#the built program has a name with format:<package>.<struct>.cfg, for example
#"conf.config.cfg"
```

Generating your configuration file with the built generator
```sh
conf.config.cfg > conf.toml #generating
conf.config.cfg -patch conf.toml #updating if conf.toml has already existed
```

## Example

### Define a struct in package conf
```go
package conf

type Config struct {
        Listen  string `cfg:"listen; :8804; netaddr; The address the server to listen"`
        MaxConn int    `cfg:"max-connection; 10000; numeric; Max number of concurrent connections"`
        Redis   struct {
                Cluster []string `cfg:"cluster; required; dialstring; The addresses of redis cluster"`
        }
}
```

## Build the configuration file generator

### First, install configo-build
```sh
go get github.com/shafreeck/configo/bin/configo-build
```

### Then, build an executable program basing on your package and struct
```sh
configo-build ./conf.Config
```
or use the absolute path
```sh
configo-build github.com/shafreeck/configo/example/conf.Config
```
### Finally, use the built program to generate a toml
```
conf.config.cfg > conf.toml
```
and you can patch you toml file if it is already existed
```
conf.config.cfg -patch conf.toml
```

Output

```toml
#type:        string
#rules:       netaddr
#description: The address the server to listen
#default:     :8804
#listen = ":8804"

#type:        int
#rules:       numeric
#description: Max number of concurrent connections
#default:     10000
#max-connection = 10000

[redis]

#type:        []string
#rules:       dialstring
#description: The addresses of redis cluster
#required
cluster = []
```
