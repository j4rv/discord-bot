package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"net/url"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/shlex"
	"github.com/jessevdk/go-flags"
	"github.com/skip2/go-qrcode"
)

// ==================== STRINGS ====================

func parseCommandArgs(opts any, input string) error {
	args, err := shlex.Split(input)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		args = args[1:]
	}

	_, err = flags.ParseArgs(opts, args)
	return err
}

// if leftToRight is false, it will make the table topToBottom
func formatInColumns(items []string, columns int, leftToRight bool) string {
	if columns <= 0 {
		return ""
	}

	rows := (len(items) + columns - 1) / columns // ceiling division

	colData := make([][]string, columns)
	if leftToRight {
		// Fill row by row (left to right)
		for i, item := range items {
			col := i % columns
			colData[col] = append(colData[col], item)
		}
	} else {
		// Fill column by column (top to bottom)
		for i, item := range items {
			col := i / rows
			colData[col] = append(colData[col], item)
		}
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

func truncateString(s string, n int) string {
	if len(s) > n {
		return s[:n] + "…"
	}
	return s
}

var badEmbedDomainReplacements = map[*regexp.Regexp]string{
	regexp.MustCompile(`\b(?:https?://)?(?:www\.)?(?:twitter|x)\.com\b`): "https://fxtwitter.com",
	regexp.MustCompile(`\b(?:https?://)?(?:www\.)?pixiv\.net\b`):         "https://phixiv.net",
	regexp.MustCompile(`\b(?:https?://)?(?:www\.)?bilibili\.com\b`):      "https://vxbilibili.com",
}

var trackingParamsByDomain = map[string][]string{
	"twitter.com":       {"t", "s"},
	"x.com":             {"t", "s"},
	"youtu.be":          {"si"},
	"youtube.com":       {"si"},
	"music.youtube.com": {"si"},
	"open.spotify.com":  {"si"},
	"bilibili.com":      {"vd_source", "spm_id_from", "share_source"},
	"*":                 {"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "fbclid", "gclid", "igshid"},
}

func sanitizeURL(raw string) string {
	inputUrl, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	// clean original host for trackingParamsByDomain
	ogHost := strings.ToLower(inputUrl.Host)
	ogHost = strings.TrimPrefix(ogHost, "www.")
	if colonIdx := strings.Index(ogHost, ":"); colonIdx != -1 {
		ogHost = ogHost[:colonIdx]
	}

	// domain changes (x.com to fxtwitter.com for example)
	for rgx, rpl := range badEmbedDomainReplacements {
		if rgx.MatchString(inputUrl.Host) {
			inputUrl.Scheme = "https"
			inputUrl.Host = strings.TrimPrefix(strings.TrimPrefix(rpl, "https://"), "http://")
			break
		}
	}

	// remove the query param trackers
	paramsToRemove := make([]string, 0)
	if perDomain, ok := trackingParamsByDomain[strings.ToLower(ogHost)]; ok {
		paramsToRemove = append(paramsToRemove, perDomain...)
	}
	paramsToRemove = append(paramsToRemove, trackingParamsByDomain["*"]...)

	q := inputUrl.Query()
	for _, p := range paramsToRemove {
		q.Del(p)
	}
	inputUrl.RawQuery = q.Encode()

	return inputUrl.String()
}

func cleanMessageContent(content string) string {
	urlRegex := regexp.MustCompile(`https?://[^\s~|<>]+`)
	return urlRegex.ReplaceAllStringFunc(content, sanitizeURL)
}

// ==================== MATH ====================

func divideToFloat(a, b int) float64 {
	return float64(a) / float64(b)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
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

// ==================== CHANNELS ====================

func channelBelongsToGuild(ds *discordgo.Session, channelID, guildID string) bool {
	if channelID == "" {
		return true
	}
	channel, err := ds.State.Channel(channelID)
	if err != nil {
		channel, err = ds.Channel(channelID)
		if err != nil {
			serverNotifyIfErr("channelBelongsToGuild", err, guildID, ds)
			return false
		}
	}
	return channel.GuildID == guildID
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

	// Index 0 is transparent cause of discord's app glitching otherwise
	palette := []color.Color{color.Transparent, color.White, color.Black}
	img := image.NewPaletted(image.Rect(0, 0, size, size), palette)

	// Fill white background
	for i := range img.Pix {
		img.Pix[i] = 1
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
						img.Pix[row+dx] = 2 // Black
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
