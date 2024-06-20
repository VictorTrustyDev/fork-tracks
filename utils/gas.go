package utils

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

func GetRequiredFeeFromError(err error) (requiredFee uint, ok bool) {
	if err == nil {
		return
	}

	/**
	Sample error message: error code: '13' msg: 'insufficient fees; got: 550amf required: 20419amf: insufficient fee'
	*/

	errMsg := err.Error()
	if !strings.Contains(errMsg, "insufficient fee") {
		return
	}

	matches := regexp.MustCompile(`:\s(\d+)[a-zA-Z]+`).FindAllStringSubmatch(errMsg, -1)
	if len(matches) != 2 || len(matches[0]) != 2 || len(matches[1]) != 2 {
		return
	}

	fmt.Println(matches)

	got, err := strconv.ParseUint(matches[0][1], 10, 64)
	if err != nil || got < 1 {
		return
	}

	required, err := strconv.ParseUint(matches[1][1], 10, 64)
	if err != nil || required < 1 {
		return
	}

	if got >= required {
		return
	}

	return uint(required), true
}
