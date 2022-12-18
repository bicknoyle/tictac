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

type Player struct {
	id    string
	sigil string
	cpu   bool
}

func (player *Player) Name() string {
	name := "Player " + player.id

	if player.cpu {
		name += " (cpu)"
	}

	return name
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

func GetEmpty(board *Board) [][]int {
	var empty [][]int
	size := len(board.grid)
	for row := 0; row < size; row++ {
		for col := 0; col < size; col++ {
			if board.grid[row][col] == "" {
				empty = append(empty, []int{row, col})
			}
		}
	}

	return empty
}

// Pick play coordinates for the CPU player based on a variety of tactics
func CpuPick(board *Board, sigil string) []int {
	logger := log.New(os.Stderr, "Cpu Tactic: ", 0)
	size := len(board.grid)
	var opSigil string
	if sigil == "X" {
		opSigil = "O"
	} else {
		opSigil = "X"
	}

	// Tactic 0: Take a random corner
	if board.turns == 0 {
		logger.Println("random corner")
		starts := [][]int{{0, 0}, {0, size - 1}, {size - 1, 0}, {size - 1, size - 1}, {size / 2, size / 2}}
		coords := starts[rand.Intn(len(starts)-1)]

		return []int{coords[0], coords[1]}
	}

	// Tactic 1: Take a corner on first move if player didn't, else take center
	if board.turns == 1 {
		if board.grid[0][0] == "" && board.grid[0][size-1] == "" &&
			board.grid[size-1][0] == "" && board.grid[size-1][size-1] == "" {
			logger.Println("take corner")
			// pick a corner neighboring opp's play. This only works on a 3x3 board
			if board.grid[0][1] != "" || board.grid[1][0] != "" {
				return []int{0, 0}
			}
			return []int{size - 1, size - 1}
		} else {
			logger.Println("take center")
			center := size / 2
			return []int{center, center}
		}
	}

	// Tactic 2: if CPU needs one to win, take that spot
	oCounts := MissingCounts(board, sigil)
	if len(oCounts[1]) > 0 {
		logger.Println("winning move")
		return []int{oCounts[1][0][0][0], oCounts[1][0][0][1]}
	}

	// Tactic 2: if player needs one to win, block that spot
	xCounts := MissingCounts(board, opSigil)
	if len(xCounts[1]) > 0 {
		logger.Println("blocker")
		return []int{xCounts[1][0][0][0], xCounts[1][0][0][1]}
	}

	// Tactic 3: take a spot that will put at least 2 in a row
	// TODO: if there are 2 turns and center is not take, take center
	if len(oCounts[2]) > 0 {
		logger.Println("near win")
		row, col := RandomPair(oCounts[2])
		return []int{row, col}
	}

	// Tactic 4: take a spot in a winnable lane
	// TODO: is this situation possible on 3x3 grid? Other tactics will usually
	// take precendence
	if len(oCounts[3]) > 0 {
		logger.Println("winnable lane")
		return []int{oCounts[3][0][0][0], oCounts[3][0][0][1]}
	}

	// Last Tactic: random empty space
	logger.Println("random")
	empty := GetEmpty(board)
	if len(empty) == 1 {
		return empty[0]
	}
	return empty[rand.Intn(len(empty)-1)]
}

func RandomPair(lanes [][][]int) (int, int) {
	var lane [][]int
	if len(lanes) == 1 {
		lane = lanes[0]
	} else {
		lane = lanes[rand.Intn(len(lanes)-1)]
	}

	var pair []int
	if len(lane) == 2 {
		pair = lane[0]
	} else {
		pair = lane[rand.Intn(len(lane)-1)]
	}

	return pair[0], pair[1]
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

func MissingCounts(board *Board, sigil string) [][][][]int {
	size := len(board.grid)
	counts := make([][][][]int, size+1)
OUTER:
	for _, set := range board.checks {
		var empties [][]int
		for _, pair := range set {
			row := pair[0]
			col := pair[1]
			if board.grid[row][col] == "" {
				empties = append(empties, []int{row, col})
			} else if board.grid[row][col] != sigil {
				continue OUTER
			}
		}
		if len(empties) > 0 {
			counts[len(empties)] = append(counts[len(empties)], empties)
		}
	}

	return counts
}

func main() {
	// TODO: assume 3x3 grid
	const SIZE = 3
	const MAX_TURNS = SIZE * SIZE

	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	firstPlayer := Player{
		id:    "1",
		sigil: "X",
		cpu:   false, // TODO: make configurable
	}

	secondPlayer := Player{
		id:    "2",
		sigil: "O",
		cpu:   true,
	}

	// TODO: associate with a player
	var wins int
	var draws int

GAMELOOP:
	for {
		board := MakeBoard(SIZE)

		fmt.Printf("New game, %v goes first...\n", firstPlayer.Name())
		PrintBoard(board)

		for {
			var currentPlayer Player
			if board.turns%2 == 0 {
				currentPlayer = firstPlayer
			} else {
				currentPlayer = secondPlayer
			}

			var coords []int
			var error error
			if currentPlayer.cpu {
				coords = CpuPick(board, currentPlayer.sigil)
				fmt.Printf("%v picked %v %v\n", currentPlayer.Name(), coords[0], coords[1])
			} else {
				fmt.Print(currentPlayer.Name() + "> ")
				coords, error = GetInput(reader, board)
			}

			if error == nil {
				board.grid[coords[0]][coords[1]] = currentPlayer.sigil
				board.turns += 1
			} else if error.Error() == "exit" {
				fmt.Printf("%v is a quitter, cya\n", currentPlayer.Name())
				break GAMELOOP
			} else {
				fmt.Println(error)
				continue
			}

			PrintBoard(board)

			if EvalBoard(board, currentPlayer.sigil) {
				wins++
				PrintResult(currentPlayer.Name() + " wins!")
				break
			} else if board.turns >= MAX_TURNS {
				draws++
				PrintResult("cat's game")
				break
			}
		}

		tmpPlayer := firstPlayer
		firstPlayer = secondPlayer
		secondPlayer = tmpPlayer

		fmt.Printf("wins: %v, draws: %v\n", wins, draws)

		if wins > 0 {
			fmt.Println("Whoa how did you win?")
			break GAMELOOP
		} else if draws == 50 {
			fmt.Println("I've seen enough...")
			break GAMELOOP
		}

		fmt.Print("Starting a new game...\n\n")
	}
}
