package utils

import "strconv"

func Atoi(value string) int {
	num, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return num
}
