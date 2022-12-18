package main

import (
	"testing"
)

func TestMissingCounts(t *testing.T) {
	board := MakeBoard(3)

	want := len(board.Checks)
	counts := MissingCounts(board, "O")
	have := len(counts[3])

	if want != have {
		t.Fatalf("MissingCounts(), all checks have count == %v, have: %v", want, have)
	}

	board.Grid = [][]string{
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

	board.Grid = [][]string{
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

func TestPlayerName(t *testing.T) {
	player := Player{
		Id:    "1",
		Sigil: "X",
	}

	name := player.Name()
	want := "Player 1"

	if name != want {
		t.Fatalf("Player.Name() = %v, want %v", name, want)
	}

	player = Player{
		Id:    "2",
		Sigil: "O",
		Cpu:   true,
	}

	name = player.Name()
	want = "Player 2 (cpu)"

	if name != want {
		t.Fatalf("Player.Name() = %v, want %v", name, want)
	}
}
