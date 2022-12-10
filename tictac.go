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

// Make a slice of slices to represent the game grid
func MakeGrid(n int) [][]string {
	grid := make([][]string, n)
	for i := 0; i < n; i++ {
		grid[i] = make([]string, n)
	}

	return grid
}

// Calculate groups of coordinates to eval for win checks
func MakeCheckCoords(n int) [][][]int {
	var coords [][][]int

	// horizontal and vertical coords
	for row := 0; row < n; row++ {
		var set_horizontal [][]int
		var set_vertical [][]int
		for col := 0; col < n; col++ {
			set_horizontal = append(set_horizontal, []int{row, col})
			set_vertical = append(set_vertical, []int{col, row})
		}
		coords = append(coords, set_horizontal, set_vertical)
	}

	// ltr diag
	row := 0
	col := 0
	var ltr_set [][]int
	for row < n {
		ltr_set = append(ltr_set, []int{row, col})
		row++
		col++
	}
	coords = append(coords, ltr_set)

	// rtl diag
	row = n - 1
	col = n - 1
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
func PrintGrid(grid [][]string) {
	for _, row := range grid {
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

func GetInput(reader *bufio.Reader, grid [][]string) ([]int, error) {
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

	if grid[y][x] != "" {
		return nil, errors.New("coord taken")
	}

	return []int{y, x}, nil
}

func CpuPick(n int, grid [][]string, turns int, check_coords [][][]int) []int {
	if turns == 1 && n%2 == 1 {
		// if first turn and odd sized grid...
		center := n / 2
		if grid[center][center] == "" {
			// if center is open, take it
			log.Println("Cpu Tactic: take center")
			return []int{center, center}
		} else {
			// take a corner
			log.Println("Cpu Tactic: take corner")
			return []int{0, 0}
		}
	} else {
		winner := MissingOne(check_coords, grid, "O")
		if winner != nil {
			// winning move
			log.Println("Cpu Tactic: winning move")
			return []int{winner[0], winner[1]}
		}

		blocker := MissingOne(check_coords, grid, "X")
		if blocker != nil {
			// block opponent
			log.Println("Cpu Tactic: blocker")
			return []int{blocker[0], blocker[1]}
		}

		// eh, just pick something random
		log.Println("Cpu Tactic: random")
		for {
			row := rand.Intn(n)
			col := rand.Intn(n)
			if grid[row][col] == "" {
				return []int{row, col}
			}
		}
	}
}

func EvalGrid(check_coords [][][]int, grid [][]string, sigil string) bool {
	need := len(grid)

	for _, set := range check_coords {
		found := 0
		for _, pair := range set {
			row := pair[0]
			col := pair[1]
			if grid[row][col] == sigil {
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
func MissingOne(check_coords [][][]int, grid [][]string, sigil string) []int {
	for _, set := range check_coords {
		var last_empty []int
		found := 0
		for _, pair := range set {
			row := pair[0]
			col := pair[1]
			if grid[row][col] == sigil {
				found++
			} else if grid[row][col] == "" {
				last_empty = []int{row, col}
			}
		}
		if found == len(grid)-1 && len(last_empty) > 0 {
			return last_empty
		}
	}

	return nil
}

func main() {
	const N = 3

	grid := MakeGrid(N)
	check_coords := MakeCheckCoords(N)

	PrintGrid(grid)

	rand.Seed(time.Now().UnixNano())
	reader := bufio.NewReader(os.Stdin)
	turns := 0
	for {
		var player string
		var sigil string
		if turns%2 == 0 {
			player = "1"
			sigil = "X"
		} else {
			player = "2"
			sigil = "O"
		}

		var coords []int
		var error error
		if player == "2" {
			coords = CpuPick(N, grid, turns, check_coords)
			fmt.Printf("Player %v picked %v %v\n", player, coords[0], coords[1])
		} else {
			fmt.Print("Player " + player + "> ")
			coords, error = GetInput(reader, grid)
		}

		if error == nil {
			grid[coords[0]][coords[1]] = sigil
			turns += 1
		} else if error.Error() == "exit" {
			fmt.Println("Player " + player + " is a quitter, cya")
			break
		} else {
			fmt.Println(error)
			continue
		}

		PrintGrid(grid)

		if EvalGrid(check_coords, grid, sigil) {
			PrintResult("Player " + player + " wins!")
			break
		} else if turns >= 9 {
			PrintResult("cat's game")
			break
		}
	}
}
