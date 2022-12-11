package main

import (
	"testing"
)

func TestMissingCounts(t *testing.T) {
	board := MakeBoard(3)

	want := len(board.checks)
	counts := MissingCounts(board, "O")
	have := len(counts[3])

	if want != have {
		t.Fatalf("MissingCounts(), all checks have count == %v, have: %v", want, have)
	}

	board.grid = [][]string{
		{"X", "", ""},
		{"", "", ""},
		{"", "", ""},
	}

	counts = MissingCounts(board, "X")

	if len(counts[2]) != 3 {
		t.Fatalf("MissingCounts(), exepected 3 sets where 2 chars are missing")
	}

	if len(counts[3]) != 5 {
		t.Fatalf("MissingCounts(), exepected 5 sets where 3 chars are missing")
	}

	board.grid = [][]string{
		{"X", "", ""},
		{"", "O", ""},
		{"", "", ""},
	}

	counts = MissingCounts(board, "X")

	if len(counts[2]) != 2 {
		t.Fatalf("MissingCounts(), exepected 2 sets where 2 chars are missing")
	}

	if len(counts[3]) != 2 {
		t.Fatalf("MissingCounts(), exepected 2 sets where 3 chars are missing")
	}

	counts = MissingCounts(board, "O")

	if len(counts[2]) != 3 {
		t.Fatalf("MissingCounts(), exepected 3 sets where 2 chars are missing")
	}
}