package utils

import (
	"errors"
	"regexp"
)

func Extract_string(pattern string, str_input string) (string, error) {
	var res string
	var err error
	rgx := regexp.MustCompile(pattern)
	match := rgx.FindStringSubmatch(str_input)

	if match == nil {
		err = errors.New("Pattern not found")
	} else {
		res = match[1]
	}

	return res, err
}

func Extract_bytes(pattern string, bytes_input []byte) ([]byte, error) {
	var res []byte
	var err error
	rgx := regexp.MustCompile(pattern)
	match := rgx.FindSubmatch(bytes_input)

	if match == nil {
		err = errors.New("Pattern not found")
	} else {
		res = match[1]
	}

	return res, err
}
