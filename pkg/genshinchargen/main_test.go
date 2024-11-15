package genshinchargen

import "testing"

func TestNewChar(t *testing.T) {
	char := NewChar("Gold", 0)
	t.Log(char.PrettyString())
	char = NewChar("Gold", 1)
	t.Log(char.PrettyString())
	char = NewChar("Gold", 2)
	t.Log(char.PrettyString())
	char = NewChar("Silver", 0)
	t.Log(char.PrettyString())
	char = NewChar("Alice", 0)
	t.Log(char.PrettyString())
	char = NewChar("Timmie", 0)
	t.Log(char.PrettyString())
	char = NewChar("Timmie", 1)
	t.Log(char.PrettyString())
	char = NewChar("Timmie", 2)
	t.Log(char.PrettyString())
	char = NewChar("Timmie", 3)
	t.Log(char.PrettyString())
	char = NewChar("Timmie", 4)
	t.Log(char.PrettyString())
}
