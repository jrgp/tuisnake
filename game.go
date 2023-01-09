package main

import (
	"errors"
	"github.com/gammazero/deque"
	"math/rand"
	"time"
)

type Type int

const (
	SNAKE Type = 1
	FOOD  Type = 2
)

type Direction int

const (
	RIGHT = 1
	LEFT  = 2
	UP    = 3
	DOWN  = 4
)

type Cell struct {
	Type Type
}

type Cord struct {
	x, y int
}

type Snake struct {
	board     map[Cord]Cell
	direction Direction
	tailQueue deque.Deque[Cord]
}

func (s *Snake) createFood() {
	for i := 0; i < 100; i++ {
		rand.Seed(time.Now().UnixNano())
		x := rand.Intn(WIDTH - 1)
		rand.Seed(time.Now().UnixNano())
		y := rand.Intn(HEIGHT - 1)
		cord := Cord{x: x, y: y}
		if _, ok := s.board[cord]; !ok {
			s.board[cord] = Cell{Type: FOOD}
			break
		}
	}
}

func (s *Snake) createSnake(cord Cord) {
	s.tailQueue.PushFront(cord)
	s.board[cord] = Cell{Type: SNAKE}
}

func (s *Snake) Reset() {
	// Wipe queue
	s.tailQueue.Clear()

	// Wipe Board
	s.board = map[Cord]Cell{}

	// Set default direction
	s.direction = RIGHT

	// Create a small 3-length snake pointing right
	// in the middle
	s.createSnake(Cord{x: WIDTH/2 - 1, y: HEIGHT / 2})
	s.createSnake(Cord{x: WIDTH / 2, y: HEIGHT / 2})
	s.createSnake(Cord{x: WIDTH/2 + 1, y: HEIGHT / 2})

	// Create a food
	s.createFood()
}

func (s *Snake) UpdateState() error {

	// Add new to front.
	next := s.tailQueue.At(0)
	previous := s.tailQueue.At(1)

	switch s.direction {
	case LEFT:
		next.x--
	case RIGHT:
		next.x++
	case DOWN:
		next.y++
	case UP:
		next.y--
	}

	// Can't change direction to trail. Sanity-check; this should be disallowed
	if next.x == previous.x && next.y == previous.y {
		return errors.New("can't go backwards")
	}

	// Boundary checks
	if next.x < 0 || next.y < 0 || next.x >= WIDTH-1 || next.y >= HEIGHT-1 {
		return errors.New("hit wall")
	}

	if cell, ok := s.board[next]; ok {
		switch cell.Type {
		case SNAKE:
			// Eat self checks
			return errors.New("ate self")
		case FOOD:
			// A food item? Don't pop last elem (get longer) and spawn new food
			s.createFood()
		}

	} else {
		// Kill final position (don't get longer)
		oldest := s.tailQueue.PopBack()
		delete(s.board, oldest)
	}

	// Good. Plant new one on board.
	s.createSnake(next)

	return nil
}

func (s *Snake) ChangeDirection(chosen Direction) {
	next := s.tailQueue.At(0)
	previous := s.tailQueue.At(1)

	switch chosen {
	case LEFT:
		next.x--
	case RIGHT:
		next.x++
	case DOWN:
		next.y++
	case UP:
		next.y--
	}

	// Can't change direction to trail
	if next.x == previous.x && next.y == previous.y {
		// Ignore
		return
	}

	s.direction = chosen
}
