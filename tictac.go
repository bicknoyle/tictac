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

	// top to bottom diag
	row := 0
	col := 0
	diag := [][]int{}
	for row < size {
		diag = append(diag, []int{row, col})
		row++
		col++
	}
	coords = append(coords, diag)

	// bottom to top diag
	row = size - 1
	col = 0
	diag = [][]int{}
	for row > -1 {
		diag = append(diag, []int{row, col})
		row--
		col++
	}
	coords = append(coords, diag)

	return coords
}

// Print out the game grid
func PrintBoard(board *Board) {
	for _, row := range board.grid {
		fmt.Print("[")
		for i, c := range row {
			if c == "" {
				fmt.Print("_")
			} else {
				fmt.Print(c)
			}
			if i < len(board.grid)-1 {
				fmt.Print("][")
			}
		}
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
	}

	if match, _ := regexp.MatchString("^\\d+ \\d+", text); !match {
		return nil, errors.New("bad input")
	}

	coords := strings.Fields(text)
	y, _ := strconv.Atoi(coords[0])
	x, _ := strconv.Atoi(coords[1])

	size := len(board.grid)
	if x >= size || y >= size {
		return nil, errors.New("out of bounds")
	}

	if board.grid[y][x] != "" {
		return nil, errors.New("coord taken")
	}

	return []int{y, x}, nil
}

func CpuPick(board *Board) []int {
	size := len(board.grid)
	logger := log.New(os.Stderr, "Cpu Tactic: ", 0)

	if board.turns == 1 && size%2 == 1 {
		// if first turn and odd sized grid...
		center := size / 2
		if board.grid[center][center] == "" {
			// if center is open, take it
			logger.Println("take center")
			return []int{center, center}
		} else {
			// take a corner
			logger.Println("take corner")
			return []int{0, 0}
		}
	}

	// TODO: no need to see if MissingOne until sigil has been played X times

	winner := MissingOne(board, "O")
	if winner != nil {
		// winning move
		logger.Println("winning move")
		return []int{winner[0], winner[1]}
	}

	blocker := MissingOne(board, "X")
	if blocker != nil {
		// block opponent
		logger.Println("blocker")
		return []int{blocker[0], blocker[1]}
	}

	// eh, just pick something random
	logger.Println("random")
	// TODO: make list of empty spaces and pick randomly from it?
	for {
		row := rand.Intn(size)
		col := rand.Intn(size)
		if board.grid[row][col] == "" {
			return []int{row, col}
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
// TODO: MissingOne can return as soon as it has seen board.turns non-blanks
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
	const SIZE = 3
	const MAX_TURNS = SIZE * SIZE

	board := MakeBoard(SIZE)

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
		} else if board.turns >= MAX_TURNS {
			PrintResult("cat's game")
			break
		}
	}
}
