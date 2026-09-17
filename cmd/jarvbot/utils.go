package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"math"
	"math/rand"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/shlex"
	"github.com/jessevdk/go-flags"
	"github.com/skip2/go-qrcode"
)

// ==================== STRINGS ====================

// If an error is returned, it will be wrapped in ``` because it will be shown to the user.
func parseCommandArgs(data any, input string) error {
	args, err := shlex.Split(input)
	if err != nil {
		return err
	}

	if len(args) > 0 {
		args = args[1:]
	}

	parser := flags.NewParser(data, flags.HelpFlag|flags.PassDoubleDash)
	if _, err = parser.ParseArgs(args); err != nil {
		return errors.New("```\n" + err.Error() + "\n```")
	}
	return nil
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
	regexp.MustCompile(`\b(?:https?://)?(?:www\.)?reddit\.com\b`):        "https://rxddit.com",
}

var trackingParamsByDomain = map[string][]string{
	"twitter.com":       {"t", "s"},
	"x.com":             {"t", "s"},
	"youtu.be":          {"si", "is"},
	"youtube.com":       {"si", "is"},
	"music.youtube.com": {"si", "is"},
	"open.spotify.com":  {"si", "is"},
	"bilibili.com":      {"vd_source", "spm_id_from", "share_source"},
	"reddit.com":        {"utm_source", "utm_medium", "utm_campaign", "utm_term", "utm_content", "context", "share_id"},
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

func clamp(value, min, max float64) float64 {
	return math.Max(min, math.Min(max, value))
}

// ==================== DATES ====================

func isAprilFools() bool {
	now := time.Now()
	return now.Month() == time.April && now.Day() == 1
}

// ==================== DISCORD ====================

func userString(u *discordgo.User) string {
	if u == nil {
		return "unknown"
	}
	return fmt.Sprintf("%s (%s)", u.DisplayName(), u.String())
}

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

func isShadowFeatureEnabled(guildID, property string) bool {
	value, _ := serverDS.getServerProperty(guildID, property)
	return value == serverPropYes
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

func deepFryImage(src image.Image, noiseAmount, saturation float64, jpegQuality, iterations int) (image.Image, error) {
	bounds := src.Bounds()
	img := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := src.At(x, y).RGBA()

			r8 := float64(r >> 8)
			g8 := float64(g >> 8)
			b8 := float64(b >> 8)

			// Saturation
			gray := 0.299*r8 + 0.587*g8 + 0.114*b8
			r8 = gray + (r8-gray)*saturation
			g8 = gray + (g8-gray)*saturation
			b8 = gray + (b8-gray)*saturation

			// Noise
			r8 += (rand.Float64()*2 - 1) * noiseAmount
			g8 += (rand.Float64()*2 - 1) * noiseAmount
			b8 += (rand.Float64()*2 - 1) * noiseAmount

			img.SetRGBA(x, y, color.RGBA{
				R: uint8(clamp(r8, 0, 255)),
				G: uint8(clamp(g8, 0, 255)),
				B: uint8(clamp(b8, 0, 255)),
				A: uint8(a >> 8),
			})
		}
	}

	// Repeated JPEG compression
	for i := 0; i < iterations; i++ {
		var buf bytes.Buffer

		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: jpegQuality}); err != nil {
			return nil, err
		}

		decoded, err := jpeg.Decode(&buf)
		if err != nil {
			return nil, err
		}

		img = image.NewRGBA(decoded.Bounds())
		draw.Draw(img, img.Bounds(), decoded, decoded.Bounds().Min, draw.Src)
	}

	return img, nil
}

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
