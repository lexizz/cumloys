package utils

import (
	"strconv"
)

func CheckNumberOrder(numberOrder string) bool {
	reversedNums := reverse(numberOrder)

	sum := 0
	flag := true
	for _, char := range reversedNums {
		num, err := strconv.Atoi(string(char))
		if err != nil {
			return false
		}

		if !flag {
			sum += (2 * num) % 9
			flag = true
			continue
		}

		sum += num
		flag = false
	}

	return sum%10 == 0
}

func reverse(str string) string {
	var result string

	for _, v := range str {
		result = string(v) + result
	}

	return result
}
