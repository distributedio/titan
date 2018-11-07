package configo

import (
	"io/ioutil"
)

//Load toml file and unmarshal to v, v shoud be a pointer
func Load(file string, v interface{}) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return Unmarshal(b, v)
}

//Dump the object to file in toml format
func Dump(file string, v interface{}) error {
	b, err := Marshal(v)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(file, b, 0644)
}

//Update an exist file
func Update(file string, v interface{}) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	out, err := Patch(b, v)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file, out, 0644)
}
