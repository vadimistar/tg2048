package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

const fieldSize = 4
const startSquareVal = 2

var field = make([][]int, fieldSize)
var score = 0

func showField() {
	for i := 0; i < fieldSize; i++ {
		for j := 0; j < fieldSize; j++ {
			if field[i][j] != 0 {
				fmt.Printf("|%5d|", field[i][j])
			} else {
				fmt.Print("|     |")
			}
		}
		fmt.Println()
	}
}

type GameOver struct{}

func (g GameOver) Error() string {
	return "Game over"
}

func fieldIsFull() bool {
	for i := 0; i < fieldSize; i++ {
		for j := 0; j < fieldSize; j++ {
			if field[i][j] == 0 {
				return false
			}
		}
	}
	return true
}

func checkOverlap(x *int, y *int) error {
	if fieldIsFull() {
		return GameOver{}
	}
	for {
		if field[*x][*y] != 0 {
			if *x == fieldSize-1 {
				*y = (*y + 1) % fieldSize
			}
			*x = (*x + 1) % fieldSize
			continue
		}
		return nil
	}
}

var rngSrc = rand.NewSource(time.Now().UnixNano())
var rng = rand.New(rngSrc)

func prepare() error {
	if _, _, e := addSquare(); e != nil {
		return e
	}
	if _, _, e := addSquare(); e != nil {
		return e
	}
	return nil
}

func addSquare() (int, int, error) {
	x := rng.Intn(fieldSize)
	y := rng.Intn(fieldSize)
	if e := checkOverlap(&x, &y); e != nil {
		return 0, 0, e
	}
	field[x][y] = startSquareVal
	return x, y, nil
}

func handle(b byte) bool {
	switch b {
	case 'a':
		left()
	case 'w':
		up()
	case 's':
		down()
	case 'd':
		right()
	default:
		return false
	}
	return true
}

func move(x int, y int, next func(int, int) (int, int), check func(int, int) bool, incr func(*int, *int)) {
	for ; check(x, y); incr(&x, &y) {
		nx, ny := next(x, y)
		if field[nx][ny] == 0 {
			field[nx][ny] = field[x][y]
			log.Printf("%d %d is moved from %d %d", nx, ny, x, y)
		} else if field[nx][ny] == field[x][y] {
			field[nx][ny] += field[x][y]
			score += field[nx][ny]
			log.Printf("%d %d combines with %d %d", nx, ny, x, y)
		} else {
			break
		}
		field[x][y] = 0
	}
}

func moveLeft(x int, y int) {
	move(x, y, func(x, y int) (int, int) {
		return x, y - 1
	}, func(_, y int) bool {
		return y > 0
	}, func(_, y *int) {
		*y--
	})
}

func moveUp(x int, y int) {
	move(x, y, func(x, y int) (int, int) {
		return x - 1, y
	}, func(x, _ int) bool {
		return x > 0
	}, func(x, _ *int) {
		*x--
	})
}

func moveDown(x int, y int) {
	move(x, y, func(x, y int) (int, int) {
		return x + 1, y
	}, func(x, _ int) bool {
		return x < fieldSize-1
	}, func(x, _ *int) {
		*x++
	})
}

func moveRight(x int, y int) {
	move(x, y, func(x, y int) (int, int) {
		return x, y + 1
	}, func(_, y int) bool {
		return y < fieldSize-1
	}, func(_, y *int) {
		*y++
	})
}

func left() {
	for x := 0; x < fieldSize; x++ {
		for y := 0; y < fieldSize; y++ {
			if field[x][y] != 0 {
				moveLeft(x, y)
			}
		}
	}
}

func up() {
	for x := 0; x < fieldSize; x++ {
		for y := 0; y < fieldSize; y++ {
			if field[x][y] != 0 {
				moveUp(x, y)
			}
		}
	}
}

func down() {
	for x := fieldSize - 1; x >= 0; x-- {
		for y := 0; y < fieldSize; y++ {
			if field[x][y] != 0 {
				moveDown(x, y)
			}
		}
	}
}

func right() {
	for x := 0; x < fieldSize; x++ {
		for y := fieldSize - 1; y >= 0; y-- {
			if field[x][y] != 0 {
				moveRight(x, y)
			}
		}
	}
}

func main() {
	for i := 0; i < fieldSize; i++ {
		field[i] = make([]int, fieldSize)
	}
	prepare()
	b := make([]byte, 256)
	for {
		fmt.Println("Score: ", score)
		showField()
		os.Stdin.Read(b)
		if handle(b[0]) {
			log.Println(string(b[0]), "was handled")
			x, y, err := addSquare()
			if err != nil {
				panic(err)
			}
			log.Printf("square at %d %d was generated", x, y)
		}
	}
}
