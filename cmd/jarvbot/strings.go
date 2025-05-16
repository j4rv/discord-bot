package main

import (
	"fmt"
	"strings"

	"github.com/jessevdk/go-flags"
)

func parseCommandArgs(opts interface{}, input string) error {
	args := strings.Fields(input)
	if len(args) > 0 {
		args = args[1:]
	}

	_, err := flags.ParseArgs(opts, args)
	return err
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
