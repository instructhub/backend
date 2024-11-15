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

func StringToUint64(str string) (uint64, error) {
	i, err := strconv.ParseInt(str, 10, 64)
	return uint64(i), err
}