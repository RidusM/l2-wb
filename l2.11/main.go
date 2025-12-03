package main

import (
	"fmt"
	"sort"
	"strings"
)

func main() {
	words := []string{"пятак", "пятка", "тяпка", "листок", "слиток", "столик", "стол"}
	result := findAnagrams(words)

	for key, group := range result {
		fmt.Printf("%q: %q\n", key, group)
	}
}

func findAnagrams(words []string) map[string][]string {
	if len(words) == 0 {
		return nil
	}

	anagrams := make(map[string][]string)
	firstWord := make(map[string]string)

	for _, word := range words {
		lowerWord := strings.ToLower(word)
		normalized := normalizeWord(lowerWord)

		if _, exists := firstWord[normalized]; !exists {
			firstWord[normalized] = lowerWord
		}
		anagrams[normalized] = append(anagrams[normalized], lowerWord)
	}

	result := make(map[string][]string)
	for normalized, group := range anagrams {
		if len(group) <= 1 {
			continue
		}
		sort.Strings(group)
		key := firstWord[normalized]
		result[key] = group
	}

	return result
}

func normalizeWord(word string) string {
	runes := []rune(word)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}