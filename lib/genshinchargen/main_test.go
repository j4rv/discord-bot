package genshinchargen

import "testing"

func TestNewChar(t *testing.T) {
	char := NewChar("Gold")
	t.Log(char)
	char = NewChar("Silver")
	t.Log(char)
	char = NewChar("Alice")
	t.Log(char)
	char = NewChar("Timmie")
	t.Log(char)
}
