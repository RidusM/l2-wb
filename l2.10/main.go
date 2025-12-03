package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

const (
	maxCapacity = 1024 * 1024
)

var monthOrder = map[string]int{
	"january": 1, "february": 2, "march": 3, "april": 4, "may": 5, "june": 6,
	"july": 7, "august": 8, "september": 9, "october": 10, "november": 11, "december": 12,
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "jun": 6,
	"jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
}

type Flags struct {
	column         int
	numeric        bool
	reverse        bool
	unique         bool
	month          bool
	ignoreTrailing bool
	check          bool
	suffix         bool
}

func main() {
	flags := parseFlags()

	files := flag.Args()

	if err := validateFlags(flags); err != nil {
		fmt.Fprintf(os.Stderr, "sort: %v\n", err)
		os.Exit(2)
	}

	lines, err := readInput(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort: %v\n", err)
		os.Exit(2)
	}

	if flags.check {
		if !isSorted(lines, flags) {
			fmt.Fprintf(os.Stderr, "sort: disorder detected\n")
			os.Exit(1)
		}
		os.Exit(0)
	}

	if flags.unique {
		lines = removeDuplicates(lines)
	}

	sortLines(lines, flags)

	for _, line := range lines {
		fmt.Println(line)
	}

	os.Exit(0)
}

func parseFlags() *Flags {
	flags := &Flags{}

	flag.IntVar(&flags.column, "k", 0, "sort by column number (1-indexed)")
	flag.BoolVar(&flags.numeric, "n", false, "sort by numeric value")
	flag.BoolVar(&flags.reverse, "r", false, "reverse sort order")
	flag.BoolVar(&flags.unique, "u", false, "output only unique lines")
	flag.BoolVar(&flags.month, "M", false, "sort by month name")
	flag.BoolVar(&flags.ignoreTrailing, "b", false, "ignore trailing blanks")
	flag.BoolVar(&flags.check, "c", false, "check if input is sorted")
	flag.BoolVar(&flags.suffix, "h", false, "sort with human-readable suffixes")

	flag.Parse()
	return flags
}

func validateFlags(flags *Flags) error {
	modes := 0
	if flags.numeric { modes++ }
	if flags.month { modes++ }
	if flags.suffix { modes++ }

	if modes > 1 {
		return fmt.Errorf("conflicting sort modes: only one of -n, -M, -h allowed")
	}
	return nil
}

func readInput(files []string) ([]string, error) {
	var lines []string

	inputs := files
	if len(inputs) == 0 {
		inputs = []string{"-"}
	}

	for _, name := range inputs {
		var (
			file *os.File
			err  error
		)

		if name == "-" {
			file = os.Stdin
		} else {
			file, err = os.Open(name)
			if err != nil {
				return nil, fmt.Errorf("cannot open %s: %w", name, err)
			}
			defer file.Close()
		}

		scanner := bufio.NewScanner(file)
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			src := "standard input"
			if name != "-" {
				src = name
			}
			return nil, fmt.Errorf("error reading %s: %w", src, err)
		}
	}

	return lines, nil
}

func getColumn(line string, n int, flags *Flags) string {
	if n == 0 {
		if flags.ignoreTrailing {
			return strings.TrimRightFunc(line, func(r rune) bool {
				return r == ' ' || r == '\t'
			})
		}
		return line
	}

	fields := strings.Split(line, "\t")
	if n > len(fields) {
		return ""
	}

	result := fields[n-1]
	if flags.ignoreTrailing {
		result = strings.TrimSpace(result)
	}
	return result
}

func sortLines(lines []string, flags *Flags) {
	sort.SliceStable(lines, func(i, j int) bool {
		a, b := lines[i], lines[j]
		less := compareLess(a, b, flags)
		if flags.reverse {
			return !less
		}
		return less
	})
}

func compareLess(a, b string, flags *Flags) bool {
	valA := getColumn(a, flags.column, flags)
	valB := getColumn(b, flags.column, flags)

	if flags.suffix {
		return compareSuffix(valA, valB)
	}
	if flags.numeric {
		return compareNumeric(valA, valB)
	}
	if flags.month {
		return compareMonth(valA, valB)
	}
	return valA < valB
}

func compareNumeric(a, b string) bool {
	numA, errA := strconv.ParseFloat(strings.TrimSpace(a), 64)
	numB, errB := strconv.ParseFloat(strings.TrimSpace(b), 64)
	if errA != nil || errB != nil {
		return a < b
	}
	return numA < numB
}

func compareMonth(a, b string) bool {
	monthA := monthOrder[strings.ToLower(strings.TrimSpace(a))]
	monthB := monthOrder[strings.ToLower(strings.TrimSpace(b))]

	if monthA == 0 && monthB == 0 {
		return a < b
	}
	return monthA < monthB
}

func compareSuffix(a, b string) bool {
	suffixA, errA := parseSuffix(a)
	suffixB, errB := parseSuffix(b)

	if errA != nil || errB != nil {
		return a < b
	}
	return suffixA < suffixB
}

func parseSuffix(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty")
	}

	multipliers := map[byte]float64{
		'K': 1024, 'M': 1024 * 1024, 'G': 1024 * 1024 * 1024,
		'T': 1.0995116e12, 'P': 1.1258999e15,
		'k': 1024, 'm': 1024 * 1024, 'g': 1024 * 1024 * 1024,
		't': 1.0995116e12, 'p': 1.1258999e15,
	}

	lastChar := s[len(s)-1]
	if mult, ok := multipliers[lastChar]; ok {
		numStr := s[:len(s)-1]
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return 0, err
		}
		return num * mult, nil
	}

	return strconv.ParseFloat(s, 64)
}

func removeDuplicates(lines []string) []string {
	seen := make(map[string]bool, len(lines))
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	return result
}

func isSorted(lines []string, flags *Flags) bool {
	for i := 1; i < len(lines); i++ {
		if compareLess(lines[i], lines[i-1], flags) {
			return false
		}
	}
	return true
}