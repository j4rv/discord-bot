package main

import (
	"context"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/heathcliff26/go-minesweeper/pkg/minesweeper"
)

var minesweeperNumberCell = []string{"🟦", "1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣"}
var minesweeperMine = "💥"

type boardConfig struct {
	Rows, Cols   int
	DefaultMines int
	MinMines     int
	MaxMines     int
	Messages     int
}

var boards = map[string]boardConfig{
	"S": {8, 13, 20, 5, 30, 1},
	"M": {14, 14, 40, 20, 60, 2},
	"L": {18, 16, 60, 40, 99, 3},
}

type minesweeperInput struct {
	MinesAmount int    `short:"m" long:"mines" default:"0" description:"The amount of mines the board will have."`
	Size        string `short:"s" long:"size" default:"S" choice:"S" choice:"M" choice:"L" description:"Board size (S, M, L)."`
}

func answerMinesweeper(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	var input minesweeperInput
	if err := parseCommandArgs(&input, mc.Content); err != nil {
		ds.ChannelMessageSend(mc.ChannelID, err.Error())
		return false
	}

	cfg, ok := boards[strings.ToUpper(input.Size)]
	if !ok {
		cfg = boards["S"]
	}

	mines := input.MinesAmount
	if mines == 0 {
		mines = cfg.DefaultMines
	}

	difficulty := minesweeper.Difficulty{
		Row:   cfg.Rows,
		Col:   cfg.Cols,
		Mines: max(min(mines, cfg.MaxMines), cfg.MinMines),
	}

	messageCount := cfg.Messages

	for _, msg := range MarkdownMinesweeperBoard(difficulty, messageCount) {
		if _, err := ds.ChannelMessageSend(mc.ChannelID, msg); err != nil {
			return false
		}
	}

	return true
}

// MarkdownMinesweeperBoard Makes a solvable game board with a 3x3 safe area
// then clicks all the cells in the safe area
// then transforms the board into a markdown where the unchecked cells are spoilered
func MarkdownMinesweeperBoard(difficulty minesweeper.Difficulty, messageCount int) []string {
	safeI := rand.IntN(difficulty.Row-2) + 1
	safeJ := rand.IntN(difficulty.Col-2) + 1

	game, err := minesweeper.NewGameSolvableWithIterations(
		difficulty,
		minesweeper.NewPos(safeI, safeJ),
		1000,
	)
	board := game.Field

	isInSafeCenter3x3 := func(i, j int) bool {
		return i >= safeI-1 && i <= safeI+1 &&
			j >= safeJ-1 && j <= safeJ+1
	}

	for x := range board {
		for y := range board[x] {
			if isInSafeCenter3x3(x, y) {
				game.CheckField(minesweeper.NewPos(x, y))
			}
		}
	}

	if messageCount < 1 {
		messageCount = 1
	}

	// Ceiling division so rows are distributed as evenly as possible.
	rowsPerMessage := (difficulty.Row + messageCount - 1) / messageCount

	var messages []string

	for start := 0; start < difficulty.Row; start += rowsPerMessage {
		end := min(start+rowsPerMessage, difficulty.Row)

		var str strings.Builder

		for _, row := range board[start:end] {
			for _, cell := range row {
				if !cell.Checked {
					str.WriteString("||")
				}

				if cell.Content == minesweeper.Mine {
					str.WriteString(minesweeperMine)
				} else {
					str.WriteString(minesweeperNumberCell[cell.Content])
				}

				if !cell.Checked {
					str.WriteString("||")
				}
			}
			str.WriteByte('\n')
		}

		if end == difficulty.Row {
			str.WriteString("Total mines: ")
			str.WriteString(strconv.Itoa(difficulty.Mines))
			if err != nil {
				str.WriteString(" (Needs random guesses!)")
			}
		}

		messages = append(messages, str.String())
	}

	return messages
}
