/*
Package configo is designed to handle configuration more easier.

Writing and parsing configuration is very easy with configo, there
is not any pain to add more keys in conf because configo can generate
the configuration file based on your struct definition.

Examples

Define struct with tags `cfg`

	type Config struct {
		Listen  string `cfg:"listen; :8804; netaddr; The address the server to listen"`
		MaxConn int    `cfg:"max-connection; 10000; numeric; Max number of concurrent connections"`
		Redis   struct {
			Cluster []string `cfg:"cluster; required; dialstring; The addresses of redis cluster"`
		}
	}

Marshal to configuration

	c := conf.Config{}

	data, err := configo.Marshal(c)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(data))

Unmarshal from file

	c := &conf.Config{}

	data, err := ioutil.ReadFile("conf/example.toml")
	if err != nil {
			log.Fatalln(err)
	}

	err = configo.Unmarshal(data, c)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(c)

Dump to file

	c := conf.Config{}

	if err := configo.Dump("dump.toml", c); err != nil {
			log.Fatalln(err)
	}
	log.Println("Dumpped to dump.toml")

The dumpped content is configuration friendly like this

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
	#required
	cluster = []

Load from file

	c := &conf.Config{}

	if err := configo.Load("conf/example.toml", c); err != nil {
			log.Fatalln(err)
	}
	log.Println(c)

Update an exist file
This is really useful when you add new member of struct and want to preserve the content
of your configuration file.

	c := conf.Config{}

	if err := configo.Update("conf/example.toml", c); err != nil {
		log.Fatalln(err)
	}
	log.Println("Update conf/example.toml with new struct")

Configo comes along with a command tool to help to generate and update configuration file
for you.

Build a generator with configo-build

	configo-build github.com/shafreeck/configo/example/conf.Config

or use relative package path if you are under configo directory

	configo-build ./example/conf.Config

Run the built program(name with format "<package>.<struct>.cfg")

	conf.config.cfg > example.toml

Update example.toml when you have updated struct conf.Configo

	conf.config.cfg -patch example.toml
*/
package configo
