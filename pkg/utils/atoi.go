package utils

import "strconv"

// Return string to int and ignore error
func Atoi(value string) int {
	num, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}
	return num
}
