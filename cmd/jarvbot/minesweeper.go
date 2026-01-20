package main

import (
	"context"
	"math/rand/v2"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/heathcliff26/go-minesweeper/pkg/minesweeper"
)

func answerMinesweeper(ds *discordgo.Session, mc *discordgo.MessageCreate, ctx context.Context) bool {
	_, err := ds.ChannelMessageSend(mc.ChannelID, MarkdownMinesweeperBoard())
	return err == nil
}

var minesweeperNumberCell = []string{"⬜", "1️⃣", "2️⃣", "3️⃣", "4️⃣", "5️⃣", "6️⃣", "7️⃣", "8️⃣"}
var minesweeperMine = "💥"

var minesweeperDifficulty = minesweeper.Difficulty{
	Name:  "JarvBot",
	Row:   8,
	Col:   13,
	Mines: 25,
}

// MarkdownMinesweeperBoard Makes a solvable game board with a 3x3 safe area
// then clicks all the cells in the safe area
// then transforms the board into a markdown where the unchecked cells are spoilered
func MarkdownMinesweeperBoard() string {
	var str strings.Builder
	safeI := rand.IntN(minesweeperDifficulty.Row-2) + 1
	safeJ := rand.IntN(minesweeperDifficulty.Col-2) + 1
	game := minesweeper.NewGameSolvable(minesweeperDifficulty, minesweeper.NewPos(safeI, safeJ))
	board := game.Field

	isInSafeCenter3x3 := func(i, j int) bool {
		return i >= safeI-1 && i <= safeI+1 && j >= safeJ-1 && j <= safeJ+1
	}

	for x, row := range board {
		for y := range row {
			if isInSafeCenter3x3(x, y) {
				game.CheckField(minesweeper.NewPos(x, y))
			}
		}
	}

	for _, row := range board {
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
		str.WriteRune('\n')
	}

	return str.String()
}
