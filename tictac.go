package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Board struct {
	grid   [][]string
	turns  int
	checks [][][]int
}

// make a game board
func MakeBoard(size int) *Board {
	var board Board

	board.grid = make([][]string, size)
	for row := 0; row < size; row++ {
		board.grid[row] = make([]string, size)
	}

	board.checks = MakeCheckCoords(size)

	return &board
}

// Calculate groups of coordinates to eval for win checks
func MakeCheckCoords(size int) [][][]int {
	var coords [][][]int

	// horizontal and vertical coords
	for row := 0; row < size; row++ {
		var set_horizontal [][]int
		var set_vertical [][]int
		for col := 0; col < size; col++ {
			set_horizontal = append(set_horizontal, []int{row, col})
			set_vertical = append(set_vertical, []int{col, row})
		}
		coords = append(coords, set_horizontal, set_vertical)
	}

	// ltr diag
	row := 0
	col := 0
	var ltr_set [][]int
	for row < size {
		ltr_set = append(ltr_set, []int{row, col})
		row++
		col++
	}
	coords = append(coords, ltr_set)

	// rtl diag
	row = size - 1
	col = size - 1
	var rtl_set [][]int
	for row > -1 {
		rtl_set = append(rtl_set, []int{row, col})
		row--
		col--
	}
	coords = append(coords, rtl_set)

	return coords
}

// Print out the game grid
func PrintBoard(board *Board) {
	for _, row := range board.grid {
		fmt.Print("[")
		row_out := make([]string, len(row))
		for i, c := range row {
			if c == "" {
				row_out[i] = "_"
			} else {
				row_out[i] = c
			}
		}
		fmt.Print(strings.Join(row_out, "]["))
		fmt.Println("]")
	}
}

func PrintResult(message string) {
	fmt.Println("Result: " + message)
}

func GetInput(reader *bufio.Reader, board *Board) ([]int, error) {
	text, _ := reader.ReadString('\n')
	// trim newline
	text = text[:len(text)-1]

	if text == "exit" || text == "quit" {
		return nil, errors.New("exit")
	} else if match, _ := regexp.MatchString("^[0-2] [0-2]$", text); !match {
		return nil, errors.New("bad input")
	}

	coords := strings.Fields(text)
	y, _ := strconv.Atoi(coords[0])
	x, _ := strconv.Atoi(coords[1])

	if board.grid[y][x] != "" {
		return nil, errors.New("coord taken")
	}

	return []int{y, x}, nil
}

func CpuPick(board *Board) []int {
	size := len(board.grid)
	if board.turns == 1 && size%2 == 1 {
		// if first turn and odd sized grid...
		center := size / 2
		if board.grid[center][center] == "" {
			// if center is open, take it
			log.Println("Cpu Tactic: take center")
			return []int{center, center}
		} else {
			// take a corner
			log.Println("Cpu Tactic: take corner")
			return []int{0, 0}
		}
	} else {
		winner := MissingOne(board, "O")
		if winner != nil {
			// winning move
			log.Println("Cpu Tactic: winning move")
			return []int{winner[0], winner[1]}
		}

		blocker := MissingOne(board, "X")
		if blocker != nil {
			// block opponent
			log.Println("Cpu Tactic: blocker")
			return []int{blocker[0], blocker[1]}
		}

		// eh, just pick something random
		log.Println("Cpu Tactic: random")
		// TODO: make list of empty spaces and pick randomly from it?
		for {
			row := rand.Intn(size)
			col := rand.Intn(size)
			if board.grid[row][col] == "" {
				return []int{row, col}
			}
		}
	}
}

func EvalBoard(board *Board, sigil string) bool {
	need := len(board.grid)

	for _, set := range board.checks {
		found := 0
		for _, pair := range set {
			row := pair[0]
			col := pair[1]
			if board.grid[row][col] == sigil {
				found++
			} else {
				break
			}
		}
		if found == need {
			return true
		}
	}

	return false
}

// find first coords where sigil needs just one more to win
func MissingOne(board *Board, sigil string) []int {
	for _, set := range board.checks {
		var last_empty []int
		found := 0
		for _, pair := range set {
			row := pair[0]
			col := pair[1]
			if board.grid[row][col] == sigil {
				found++
			} else if board.grid[row][col] == "" {
				last_empty = []int{row, col}
			}
		}
		if found == len(board.grid)-1 && len(last_empty) > 0 {
			return last_empty
		}
	}

	return nil
}

func main() {
	board := MakeBoard(3)

	PrintBoard(board)

	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)
	for {
		var player string
		var sigil string
		if board.turns%2 == 0 {
			player = "1"
			sigil = "X"
		} else {
			player = "2"
			sigil = "O"
		}

		var coords []int
		var error error
		if player == "2" {
			coords = CpuPick(board)
			fmt.Printf("Player %v picked %v %v\n", player, coords[0], coords[1])
		} else {
			fmt.Print("Player " + player + "> ")
			coords, error = GetInput(reader, board)
		}

		if error == nil {
			board.grid[coords[0]][coords[1]] = sigil
			board.turns += 1
		} else if error.Error() == "exit" {
			fmt.Println("Player " + player + " is a quitter, cya")
			break
		} else {
			fmt.Println(error)
			continue
		}

		PrintBoard(board)

		if EvalBoard(board, sigil) {
			PrintResult("Player " + player + " wins!")
			break
		} else if board.turns >= 9 {
			PrintResult("cat's game")
			break
		}
	}
}
