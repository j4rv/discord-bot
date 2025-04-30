package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

func parseCommandToMap(input string, expectedFlags []string) map[string]string {
	fs := flag.NewFlagSet("parseCommandToMap", flag.ContinueOnError)
	results := make(map[string]*string)

	// Register each expected flag dynamically
	for _, name := range expectedFlags {
		results[name] = fs.String(name, "", "dynamic flag")
	}

	// Parse args (skip the command name)
	args := strings.Fields(input)[1:]
	_ = fs.Parse(args)

	// Extract actual values into a result map
	out := make(map[string]string)
	for name, ptr := range results {
		if ptr != nil && *ptr != "" {
			out[name] = *ptr
		}
	}

	return out
}

func parsePaginatedCommand(input string) (int, error) {
	argMap := parseCommandToMap(input, []string{"page"})
	page := argMap["page"]
	if page == "" || page == "0" {
		page = "1"
	}
	pageInt, err := strconv.Atoi(page)
	if pageInt < 0 {
		return pageInt, errors.New("page number cannot be negative")
	}
	return pageInt, err
}

func formatInColumns(items []string, columns int) string {
	rows := (len(items) + columns - 1) / columns // ceiling division

	// Arrange items into columns
	colData := make([][]string, columns)
	for i := 0; i < len(items); i++ {
		col := i / rows
		colData[col] = append(colData[col], items[i])
	}

	// Determine max width for each column
	colWidths := make([]int, columns)
	for i, col := range colData {
		for _, item := range col {
			if len(item) > colWidths[i] {
				colWidths[i] = len(item)
			}
		}
	}

	// Build the string row by row
	var builder strings.Builder
	for row := 0; row < rows; row++ {
		for col := 0; col < columns; col++ {
			if row < len(colData[col]) {
				format := fmt.Sprintf("%%-%ds", colWidths[col]+2)
				builder.WriteString(fmt.Sprintf(format, colData[col][row]))
			}
		}
		builder.WriteByte('\n')
	}

	return builder.String()
}
