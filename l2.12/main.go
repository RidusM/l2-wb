package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
)

const (
	maxCapacity = 1024 * 1024
)

type Flags struct {
	after       int
	before      int
	context     int
	count       bool
	ignoreCase  bool
	invert      bool
	fixedString bool
	lineNumber  bool
	pattern     string
	regex       *regexp.Regexp
}

func main() {
	flags := parseFlags()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "grep: usage: grep [OPTIONS] PATTERN [FILE...]\n")
		os.Exit(2)
	}

	flags.pattern = flag.Arg(0)
	files := flag.Args()[1:]

	if err := compilePattern(flags); err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		os.Exit(2)
	}

	lines, err := readInput(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		os.Exit(2)
	}

	matches := findMatches(lines, flags)

	if flags.count {
		fmt.Println(len(matches))
		if len(matches) == 0 {
			os.Exit(1)
		}
		os.Exit(0)
	}

	if len(matches) == 0 {
		os.Exit(1)
	}

	linesToPrint := calculateLinesToPrint(matches, len(lines), flags)
	printLines(lines, linesToPrint, matches, flags)

	os.Exit(0)
}

func parseFlags() *Flags {
	flags := &Flags{}

	flag.IntVar(&flags.after, "A", 0, "print N lines of trailing context")
	flag.IntVar(&flags.before, "B", 0, "print N lines of leading context")
	flag.IntVar(&flags.context, "C", 0, "print N lines of context")
	flag.BoolVar(&flags.count, "c", false, "print only count of matching lines")
	flag.BoolVar(&flags.ignoreCase, "i", false, "ignore case")
	flag.BoolVar(&flags.invert, "v", false, "invert match")
	flag.BoolVar(&flags.fixedString, "F", false, "pattern is a fixed string")
	flag.BoolVar(&flags.lineNumber, "n", false, "print line numbers")

	flag.Parse()

	if flags.context > 0 {
		flags.after = flags.context
		flags.before = flags.context
	}

	return flags
}

func readInput(files []string) ([]string, error) {
	var lines []string

	if len(files) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		return lines, nil
	}

	for _, filename := range files {
		file, err := os.Open(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot open %s: %w", filename, err)
		}

		scanner := bufio.NewScanner(file)
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		file.Close()

		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("error reading %s: %w", filename, err)
		}
	}

	return lines, nil
}

func compilePattern(flags *Flags) error {
	pattern := flags.pattern

	if flags.fixedString {
		pattern = regexp.QuoteMeta(pattern)
	}

	if flags.ignoreCase {
		pattern = "(?i)" + pattern
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("invalid pattern: %v", err)
	}

	flags.regex = regex
	return nil
}

func findMatches(lines []string, flags *Flags) map[int]bool {
	matches := make(map[int]bool)

	for i, line := range lines {
		matched := flags.regex.MatchString(line)

		if flags.invert {
			matched = !matched
		}

		if matched {
			matches[i] = true
		}
	}

	return matches
}

func calculateLinesToPrint(matches map[int]bool, totalLines int, flags *Flags) map[int]bool {
	linesToPrint := make(map[int]bool)

	for matchIdx := range matches {
		linesToPrint[matchIdx] = true

		for i := matchIdx - flags.before; i < matchIdx; i++ {
			if i >= 0 {
				linesToPrint[i] = true
			}
		}

		for i := matchIdx + 1; i <= matchIdx+flags.after; i++ {
			if i < totalLines {
				linesToPrint[i] = true
			}
		}
	}

	return linesToPrint
}

func printLines(lines []string, linesToPrint map[int]bool, matches map[int]bool, flags *Flags) {
	var indices []int
	for idx := range linesToPrint {
		indices = append(indices, idx)
	}

	sort.Ints(indices)

	lastIdx := -2
	for _, idx := range indices {
		if flags.before > 0 || flags.after > 0 {
			if lastIdx >= 0 && idx > lastIdx+1 {
				fmt.Println("--")
			}
		}

		prefix := ""
		if flags.lineNumber {
			if matches[idx] {
				prefix = fmt.Sprintf("%d:", idx+1)
			} else if flags.before > 0 || flags.after > 0 {
				prefix = fmt.Sprintf("%d-", idx+1)
			}
		}

		fmt.Printf("%s%s\n", prefix, lines[idx])
		lastIdx = idx
	}
}