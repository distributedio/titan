package rule

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

type Validator interface {
	Validate(value string) error
}

//numeric range, format: (min,max),(min,max],[min,max),[min,max], no spaces in format
type vrange struct {
	min int64
	max int64

	left  bool //min value include
	right bool //max value include
}

func (n *vrange) String() string {
	format := "%s%d,%d%s"
	left := "("
	right := ")"
	if n.left {
		left = "["
	}
	if n.right {
		right = "]"
	}

	return fmt.Sprintf(format, left, n.min, n.max, right)
}

func (n *vrange) Validate(value string) error {
	v, err := strconv.ParseInt(value, 0, 64)
	if err != nil {
		prevErr := err
		d, err := time.ParseDuration(value)
		if err != nil {
			return prevErr
		}
		v = int64(d)
	}
	if n.left && v == n.min {
		return nil
	}
	if n.right && v == n.max {
		return nil
	}
	if v > n.min && v < n.max {
		return nil
	}
	return fmt.Errorf("value %s out of the range %s", value, n)
}

//regex match, format: /expression/
type regex struct {
	exp string
}

func (re *regex) Validate(value string) error {
	r, err := regexp.Compile(re.exp)
	if err != nil {
		return err
	}
	if r.MatchString(value) {
		return nil
	}
	return fmt.Errorf("value does not match /%s/", re.exp)
}

//named validator, format: name
type ValidatorFunc func(value string) error

type named struct {
	name string
}

func (n *named) Validate(value string) error {
	f, found := namedValidators[n.name]
	if !found {
		return fmt.Errorf("%s is not a validator", n.name)
	}

	return f(value)
}

func AddValidator(name string, f func(v string) error) error {
	namedValidators[name] = f
	return nil
}
