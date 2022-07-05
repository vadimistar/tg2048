package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

type point struct {
	x, y int
}

type square struct {
	point
	val int
}

const FIELD_SIZE = 4
const START_SQUARE_VAL = 2

type direction int

const (
	UP direction = iota
	RIGHT
	DOWN
	LEFT
)

var squares = make([]square, 0, 16)

func showField() {
	for i := 0; i < FIELD_SIZE; i++ {
		for j := 0; j < FIELD_SIZE; j++ {
			empty := true
			for _, s := range squares {
				if s.x == i && s.y == j {
					fmt.Printf("| %5d |", s.val)
					empty = false
					break
				}
			}
			if empty {
				fmt.Print("|       |")
			}
		}
		fmt.Println()
	}
}

var rngSrc = rand.NewSource(time.Now().UnixNano())
var rng = rand.New(rngSrc)

func prepare() {
	s1 := square{
		point: point{
			x: rng.Intn(FIELD_SIZE),
			y: rng.Intn(FIELD_SIZE),
		},
		val: START_SQUARE_VAL,
	}
	s2 := square{
		point: point{
			x: rng.Intn(FIELD_SIZE),
			y: rng.Intn(FIELD_SIZE),
		},
		val: START_SQUARE_VAL,
	}
	if s1.point == s2.point {
		s2.x = (s2.x + 1) % 4
	}
	squares = append(squares, s1, s2)
}

func main() {
	prepare()
	b := make([]byte, 16)
	for {
		showField()
		os.Stdin.Read(b)
		fmt.Println("I got the byte:", b[0])
	}
}
