package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/skip2/go-qrcode"
)

// ==================== STRINGS ====================

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

// ==================== MATH ====================

func divideToFloat(a, b int) float64 {
	return float64(a) / float64(b)
}

// ==================== ROLES ====================

func findRoleInSlice(roleID string, roles []*discordgo.Role) *discordgo.Role {
	for _, r := range roles {
		if r.ID == roleID {
			return r
		}
	}
	return nil
}

func guildRoleByName(ds *discordgo.Session, guildID string, roleName string) (*discordgo.Role, error) {
	roles, err := ds.GuildRoles(guildID)
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.Name == roleName {
			return r, nil
		}
	}
	return nil, fmt.Errorf("role with name %s not found in guild with id %s", roleName, guildID)
}

func isMemberInRole(member *discordgo.Member, roleID string) bool {
	for _, r := range member.Roles {
		if r == roleID {
			return true
		}
	}
	return false
}

// ==================== IMAGES ====================

func GenerateQRImage(data string, border int) ([]byte, error) {
	qr, err := qrcode.New(data, qrcode.Low)
	if err != nil {
		return nil, err
	}
	qr.DisableBorder = true

	matrix := qr.Bitmap()
	n := len(matrix)
	scale := 4
	size := (n + 2*border) * scale

	palette := []color.Color{color.White, color.Black}
	img := image.NewPaletted(image.Rect(0, 0, size, size), palette)

	// Fill white background
	for i := range img.Pix {
		img.Pix[i] = 0
	}

	// Draw QR modules with scaling
	for y := range n {
		for x := range n {
			if matrix[y][x] {
				startX := (x + border) * scale
				startY := (y + border) * scale
				for dy := range scale {
					row := (startY+dy)*img.Stride + startX
					for dx := 0; dx < scale; dx++ {
						img.Pix[row+dx] = 1 // Black
					}
				}
			}
		}
	}

	var buf bytes.Buffer
	err = gif.Encode(&buf, img, &gif.Options{
		NumColors: len(palette),
	})
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
