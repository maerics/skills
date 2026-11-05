// Package strutils holds small, independent string helper functions.
package strutils

import "strings"

// Reverse returns s with its characters reversed.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// CountVowels counts the vowels (a, e, i, o, u) in s.
func CountVowels(s string) int {
	n := 0
	for _, c := range strings.ToLower(s) {
		if strings.ContainsRune("aeiou", c) {
			n++
		}
	}
	return n
}

// Truncate returns the first n characters of s.
func Truncate(s string, n int) string {
	r := []rune(s)
	if n > len(r) {
		n = len(r)
	}
	return string(r[:n])
}
