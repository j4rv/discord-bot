package uwu

import (
	"math/rand"
	"strings"
	"testing"
)

func TestUwuify(t *testing.T) {
	rand.Seed(1)

	u := NewUwuifier()
	u.stutterChance = 1
	u.stretchChance = 1
	u.emoteChance = 1
	u.suffixChance = 1

	input := "Really lovely people are here"
	output := u.UwUify(input)

	t.Logf("input:  %s", input)
	t.Logf("output: %s", output)

	if output == input {
		t.Fatal("expected uwuified output to differ from input")
	}

	if !strings.Contains(output, "W") && !strings.Contains(output, "w") {
		t.Errorf("expected output to contain uwu letter replacements, got: %s", output)
	}
}

func TestUwuify2(t *testing.T) {
	input := "Really lovely people are here, and they are looking forward to having a wonderful conversation with everyone!"
	baseSeed := int64(12345)

	for i := int64(0); i < 8; i++ {
		seed := baseSeed + i
		rand.Seed(seed)

		u := NewUwuifier()
		output := u.UwUify(input)

		t.Logf("seed: %d", seed)
		t.Logf("input:  %s", input)
		t.Logf("output: %s", output)
	}
}

func TestReplaceLetters(t *testing.T) {
	u := NewUwuifier()

	tests := []struct {
		input    string
		expected string
	}{
		{"Hello world", "Hewwo wowwd"},
		{"Really", "Weawwy"},
		{"RLRL", "WWWW"},
		{"Nothing changes", "Nothing changes"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := u.replaceLetters(test.input)

			if got != test.expected {
				t.Errorf("replaceLetters(%q) = %q, expected %q", test.input, got, test.expected)
			}
		})
	}
}

func TestAddStutter(t *testing.T) {
	rand.Seed(1)

	u := NewUwuifier()
	got := u.addStutter("Hello")

	if !strings.HasSuffix(got, "Hello") {
		t.Errorf("expected stuttered string to end with Hello, got %q", got)
	}

	if got == "Hello" {
		t.Error("expected stutter to be added")
	}
}

func TestStretchRandomWord(t *testing.T) {
	rand.Seed(1)

	u := NewUwuifier()

	input := "hello cute banana"
	got := u.stretchRandomWord(input)

	t.Logf("input: %s", input)
	t.Logf("output: %s", got)

	if got == input {
		t.Error("expected a word to be stretched")
	}

	if len(got) <= len(input) {
		t.Errorf("expected output to be longer than input: %q", got)
	}
}

func TestStretchRandomWordNoCandidate(t *testing.T) {
	u := NewUwuifier()

	input := "why hmm grr"
	got := u.stretchRandomWord(input)

	if got != input {
		t.Errorf("expected %q, got %q", input, got)
	}
}
