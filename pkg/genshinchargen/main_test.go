package genshinchargen

import "testing"

func TestNewChar(t *testing.T) {
	char := NewChar("Gold", 0)
	t.Log(char)
	char = NewChar("Gold", 1)
	t.Log(char)
	char = NewChar("Gold", 2)
	t.Log(char)
	char = NewChar("Silver", 0)
	t.Log(char)
	char = NewChar("Alice", 0)
	t.Log(char)
	char = NewChar("Timmie", 0)
	t.Log(char)
}
