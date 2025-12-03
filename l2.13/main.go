package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Flags struct {
	fields    string
	delimiter string
	separated bool
}

type FieldSet struct {
	fixed      map[int]bool
	fromStart  *int 
	fromEnd    *int
	all        bool         
}

func (fs *FieldSet) contains(idx int) bool {
	if fs.all {
		return true
	}
	if fs.fixed[idx] {
		return true
	}
	if fs.fromStart != nil && idx <= *fs.fromStart {
		return true
	}
	if fs.fromEnd != nil && idx >= *fs.fromEnd {
		return true
	}
	return false
}

func main() {
	flags := parseFlags()

	if flags.fields == "" {
		fmt.Fprintf(os.Stderr, "error: -f flag is required\n")
		os.Exit(1)
	}

	delimiter := flags.delimiter
	if delimiter == "" {
		delimiter = "\t"
	}
	if len(delimiter) != 1 {
		fmt.Fprintf(os.Stderr, "error: delimiter must be exactly one character\n")
		os.Exit(1)
	}
	del := rune(delimiter[0])

	fieldSet, err := parseFields(flags.fields)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	lines, err := readInput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading input: %v\n", err)
		os.Exit(1)
	}

	for _, line := range lines {
		if flags.separated && !strings.ContainsRune(line, del) {
			continue
		}

		parts := strings.Split(line, string(del))
		var output []string

		for i := 0; i < len(parts); i++ {
			if fieldSet.contains(i) {
				output = append(output, parts[i])
			}
		}

		if len(output) > 0 {
			fmt.Println(strings.Join(output, string(del)))
		}
	}
}

func parseFlags() *Flags {
	flags := &Flags{}

	flag.StringVar(&flags.fields, "f", "", "indication of the fields (columns) that need to be displayed.")
	flag.StringVar(&flags.delimiter, "d", "", "use another separator (symbol).")
	flag.BoolVar(&flags.separated, "s", false, "if the flag is indicated, the lines without a separator are ignored (not displayed).")

	flag.Parse()

	return flags
}

func parseFields(spec string) (*FieldSet, error) {
	if spec == "" {
		return nil, fmt.Errorf("fields spec is empty")
	}

	fs := &FieldSet{
		fixed: make(map[int]bool),
	}

	parts := strings.Split(spec, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if part == "-" {
			fs.all = true
			continue
		}

		if strings.Contains(part, "-") {
			bounds := strings.Split(part, "-")
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid range: %s", part)
			}
			startStr := strings.TrimSpace(bounds[0])
			endStr := strings.TrimSpace(bounds[1])

			if startStr == "" && endStr != "" {
				end, err := strconv.Atoi(endStr)
				if err != nil || end < 1 {
					return nil, fmt.Errorf("invalid end in range: %s", endStr)
				}
				idx := end - 1
				fs.fromStart = &idx
				continue
			}

			if startStr != "" && endStr == "" {
				start, err := strconv.Atoi(startStr)
				if err != nil || start < 1 {
					return nil, fmt.Errorf("invalid start in range: %s", startStr)
				}
				idx := start - 1
				fs.fromEnd = &idx
				continue
			}

			if startStr != "" && endStr != "" {
				start, err := strconv.Atoi(startStr)
				if err != nil || start < 1 {
					return nil, fmt.Errorf("invalid start: %s", startStr)
				}
				end, err := strconv.Atoi(endStr)
				if err != nil || end < 1 {
					return nil, fmt.Errorf("invalid end: %s", endStr)
				}
				if start > end {
					return nil, fmt.Errorf("invalid range %d-%d: start > end", start, end)
				}
				for i := start; i <= end; i++ {
					fs.fixed[i-1] = true
				}
				continue
			}

			if startStr == "" && endStr == "" {
				fs.all = true
				continue
			}
		} else {
			num, err := strconv.Atoi(part)
			if err != nil || num < 1 {
				return nil, fmt.Errorf("invalid field: %s", part)
			}
			fs.fixed[num-1] = true
		}
	}

	return fs, nil
}

func readInput() ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}