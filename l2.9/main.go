package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"
)

var (
	ErrInvalidString = errors.New(`"" | invalid string format`)
)

func Unpack(s string) (string, error) {
	if s == "" {
		return "", nil
	}

	runes := []rune(s)
	
	var result strings.Builder
	result.Grow(len(s) * 2)

	escaped := false

	var prevRune rune
	hasPrevRune := false

	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '\\' && !escaped {
			escaped = true
			continue
		}

		if escaped {
			if hasPrevRune {
				result.WriteRune(prevRune)
			}
			prevRune = r
			hasPrevRune = true
			escaped = false
			continue
		}

		if unicode.IsDigit(r) {
			if !hasPrevRune {
				return "", ErrInvalidString
			}

			digitStr := string(r)
			for i+1 < len(runes) && unicode.IsDigit(runes[i+1]) {
				i++
				digitStr += string(runes[i])
			}

			count, err := strconv.Atoi(digitStr)
			if err != nil {
				return "", ErrInvalidString
			}

			result.WriteString(strings.Repeat(string(prevRune), count))
			hasPrevRune = false
			prevRune = 0
		} else {
			if hasPrevRune {
				result.WriteRune(prevRune)
			}
			prevRune = r
			hasPrevRune = true
		}
	}

	if escaped {
		return "", ErrInvalidString
	}

	if hasPrevRune {
		result.WriteRune(prevRune)
	}

	return result.String(), nil
}

func main() {
	testCases := []string{
		"a4bc2d5e",
		"abcd",
		"45",
		"",
		`qwe\4\5`,
		`qwe\45`,
	}

	for _, input := range testCases {
		result, err := Unpack(input)
		if err != nil {
			fmt.Printf("Input:  %q\n", input)
			fmt.Printf("Error:  %v\n\n", err)
		} else {
			fmt.Printf("Input:  %q\n", input)
			fmt.Printf("Output: %q\n\n", result)
		}
	}

	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			result, err := Unpack(arg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%q -> %q\n", arg, result)
		}
	}
}