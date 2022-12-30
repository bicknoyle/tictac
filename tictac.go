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
	Grid   [][]string
	Size   int
	Turns  int
	Checks [][][]int
}

func (board *Board) Set(row int, col int, sigil string) {
	board.Grid[row][col] = sigil
	board.Turns += 1
}

func (board *Board) Get(row int, col int) (string, error) {
	if row >= board.Size || col >= board.Size {
		return "", errors.New("out of bounds")
	}

	return board.Grid[row][col], nil
}

// stringify the current board state
func (board *Board) String() string {
	result := ""

	for _, row := range board.Grid {
		result += "["
		for i, c := range row {
			if c == "" {
				result += "_"
			} else {
				result += c
			}
			if i < board.Size-1 {
				result += "]["
			}
		}
		result += "]\n"
	}

	return result
}

type Player struct {
	Id    string
	Sigil string
	Cpu   bool
}

func (player *Player) Name() string {
	name := "Player " + player.Id

	if player.Cpu {
		name += " (cpu)"
	}

	return name
}

// make a game board
func MakeBoard() *Board {
	var board Board

	board.Size = 3

	board.Grid = make([][]string, board.Size)
	for row := 0; row < board.Size; row++ {
		board.Grid[row] = make([]string, board.Size)
	}

	board.Checks = MakeCheckCoords(board.Size)

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
	row, _ := strconv.Atoi(coords[0])
	col, _ := strconv.Atoi(coords[1])

	sigil, error := board.Get(row, col)

	if error != nil {
		return nil, error
	} else if sigil != "" {
		return nil, errors.New("coord taken")
	}

	return []int{row, col}, nil
}

func GetEmpty(board *Board) [][]int {
	var empty [][]int
	for row := 0; row < board.Size; row++ {
		for col := 0; col < board.Size; col++ {
			sigil, _ := board.Get(row, col)
			if sigil == "" {
				empty = append(empty, []int{row, col})
			}
		}
	}

	return empty
}

// Pick play coordinates for the CPU player based on a variety of tactics
func CpuPick(board *Board, sigil string) []int {
	logger := log.New(os.Stderr, "Cpu Tactic: ", 0)
	var opSigil string
	if sigil == "X" {
		opSigil = "O"
	} else {
		opSigil = "X"
	}

	// Tactic 0: Take a random corner
	if board.Turns == 0 {
		logger.Println("random corner")
		starts := [][]int{{0, 0}, {0, board.Size - 1}, {board.Size - 1, 0}, {board.Size - 1, board.Size - 1}, {board.Size / 2, board.Size / 2}}
		coords := starts[rand.Intn(len(starts)-1)]

		return []int{coords[0], coords[1]}
	}

	// Tactic 1: Take a corner on first move if player didn't, else take center
	if board.Turns == 1 {
		nw, _ := board.Get(0, 0)
		ne, _ := board.Get(0, board.Size-1)
		sw, _ := board.Get(board.Size-1, 0)
		se, _ := board.Get(board.Size-1, board.Size-1)
		if nw == "" && ne == "" && sw == "" && se == "" {
			logger.Println("take corner")
			// pick a corner neighboring opp's play. This only works on a 3x3 board
			neighborA, _ := board.Get(0, 1)
			neighborB, _ := board.Get(1, 0)
			if neighborA != "" || neighborB != "" {
				return []int{0, 0}
			}
			return []int{board.Size - 1, board.Size - 1}
		} else {
			logger.Println("take center")
			center := board.Size / 2
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
	for _, set := range board.Checks {
		found := 0
		for _, pair := range set {
			row := pair[0]
			col := pair[1]
			foundSigil, _ := board.Get(row, col)
			if foundSigil == sigil {
				found++
			} else {
				break
			}
		}
		if found == board.Size {
			return true
		}
	}

	return false
}

func MissingCounts(board *Board, sigil string) [][][][]int {
	counts := make([][][][]int, board.Size+1)
OUTER:
	for _, set := range board.Checks {
		var empties [][]int
		for _, pair := range set {
			row := pair[0]
			col := pair[1]
			foundSigil, _ := board.Get(row, col)
			if foundSigil == "" {
				empties = append(empties, []int{row, col})
			} else if foundSigil != sigil {
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
	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)

	firstPlayer := Player{
		Id:    "1",
		Sigil: "X",
		Cpu:   false, // TODO: make configurable
	}

	secondPlayer := Player{
		Id:    "2",
		Sigil: "O",
		Cpu:   true,
	}

	// TODO: associate with a player
	var wins int
	var draws int

GAMELOOP:
	for {
		board := MakeBoard()
		MAX_TURNS := board.Size * board.Size

		fmt.Printf("New game, %v goes first...\n", firstPlayer.Name())
		fmt.Print(board)

		for {
			var currentPlayer Player
			if board.Turns%2 == 0 {
				currentPlayer = firstPlayer
			} else {
				currentPlayer = secondPlayer
			}

			var coords []int
			var error error
			if currentPlayer.Cpu {
				coords = CpuPick(board, currentPlayer.Sigil)
				fmt.Printf("%v picked %v %v\n", currentPlayer.Name(), coords[0], coords[1])
			} else {
				fmt.Print(currentPlayer.Name() + "> ")
				coords, error = GetInput(reader, board)
			}

			if error == nil {
				board.Set(coords[0], coords[1], currentPlayer.Sigil)
			} else if error.Error() == "exit" {
				fmt.Printf("%v is a quitter, cya\n", currentPlayer.Name())
				break GAMELOOP
			} else {
				fmt.Println(error)
				continue
			}

			fmt.Print(board)

			if EvalBoard(board, currentPlayer.Sigil) {
				wins++
				PrintResult(currentPlayer.Name() + " wins!")
				break
			} else if board.Turns >= MAX_TURNS {
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
