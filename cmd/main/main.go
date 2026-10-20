package main

import (
	"fmt"
)

func isBalanced(s string) bool {
	var stack []rune = make([]rune, 0)

	var matching map[rune]rune = map[rune]rune{')': '(', '}': '{', ']': '['}
	for _, c := range s {
		if c == '(' || c == '{' || c == '[' {
			stack = append(stack, c)
		} else {
			if len(stack) == 0 {
				return false
			}
			if stack[len(stack)-1] == matching[c] {
				fmt.Println(stack)
				stack = stack[:len(stack)-1]
				fmt.Println(stack)
			} else {
				return false
			}
		}

	}
	if len(stack) > 0 {
		return false
	}
	return true
}

func main() {

	nums := [...]int{1, 2, 3, 4}
	fmt.Println(nums[3])
}
