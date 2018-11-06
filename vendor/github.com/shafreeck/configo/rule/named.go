package rule

import (
	"errors"
	"net"
	"net/url"
	"strconv"
)

func init() {
	namedValidators = make(map[string]ValidatorFunc)
	namedValidators["netaddr"] = ValidateNetAddr
	namedValidators["url"] = ValidateURL
	namedValidators["nonempty"] = ValidateNonempty
	namedValidators["dialstring"] = ValidateDialString
	namedValidators["boolean"] = ValidateBoolean
	namedValidators["numeric"] = ValidateNumeric
	namedValidators["printableascii"] = ValidatePrintableASCII
	namedValidators["path"] = ValidatePath
}

var namedValidators map[string]ValidatorFunc

func ValidateNetAddr(value string) error {
	_, p, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}

	_, err = strconv.ParseInt(p, 0, 64)
	if err != nil {
		return err
	}
	return nil
}

//A URL represents a parsed URL (technically, a URI reference). The general form represented is:
//scheme://[userinfo@]host/path[?query][#fragment]
func ValidateURL(value string) error {
	if _, err := url.Parse(value); err != nil {
		return err
	}
	return nil
}

func ValidateNonempty(value string) error {
	//empty string
	if value == "" {
		return errors.New("value should not be empty")
	}
	//empty array
	if value == "[]" {
		return errors.New("array should not be empty")
	}
	return nil
}

func ValidateDialString(value string) error {
	h, p, err := net.SplitHostPort(value)
	if err != nil {
		return err
	}

	_, err = strconv.ParseInt(p, 0, 64)
	if err != nil {
		return err
	}

	if h == "" {
		return errors.New("host should not be empty")
	}
	return nil
}
func ValidateBoolean(value string) error {
	if value == "true" || value == "false" {
		return nil
	}
	return errors.New("value should be true or false")
}
func ValidateNumeric(value string) error {
	if rxNumeric.MatchString(value) {
		return nil
	}
	return errors.New("value should be numeric")
}
func ValidatePrintableASCII(value string) error {
	if rxPrintableASCII.MatchString(value) {
		return nil
	}
	return errors.New("value should be ascii text")
}
func ValidatePath(value string) error {
	if rxUnixPath.MatchString(value) {
		return nil
	}
	return errors.New("value should be unix path")
}
