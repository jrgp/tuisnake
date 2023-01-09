package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gammazero/deque"
	"github.com/gdamore/tcell"
)

type Type int

const (
	EMPTY Type = 1
	SNAKE      = 2
	FOOD       = 3
)

type Direction int

const (
	RIGHT = 1
	LEFT  = 2
	UP    = 3
	DOWN  = 4
)

const (
	WIDTH  = 30
	HEIGHT = 30
)

type Cell struct {
	Type Type
	Age  int
}

type Cord struct {
	x, y int
}

// TODO make this a map[Cord]Cell{}
var board = [WIDTH][HEIGHT]Cell{}

var (
	direction Direction
	tailQueue deque.Deque[Cord]
)

func initState() {
	// Create board
	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			board[i][j] = Cell{Type: EMPTY}
		}
	}

	// Set default direction
	direction = RIGHT

	// Create a small 3-length snake pointing right
	// in the middle
	createSnake(WIDTH/2-1, HEIGHT/2)
	createSnake(WIDTH/2, HEIGHT/2)
	createSnake(WIDTH/2+1, HEIGHT/2)

	// Create a food
	createFood()
}

// TODO: take in a Cord
func createSnake(x, y int) {
	tailQueue.PushFront(Cord{x, y})
	board[x][y] = Cell{Type: SNAKE}
	//log.Printf("setting %v/%v to snake", x, y)
}

func createFood() {
	for i := 0; i < 100; i++ {
		rand.Seed(time.Now().UnixNano())
		x := rand.Intn(WIDTH - 1)
		rand.Seed(time.Now().UnixNano())
		y := rand.Intn(HEIGHT - 1)
		if board[x][y].Type == EMPTY {
			board[x][y].Type = FOOD
			//log.Printf("setting %v/%v to food", x, y)
			break
		}
	}
}

// Try to move TIP one cell to the right
// Change the oldest CELL tail to empty
func frame() error {

	// Add new to front.
	newest := tailQueue.At(0)
	previous := tailQueue.At(1)

	next := Cord{x: newest.x, y: newest.y}

	switch direction {
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

	switch board[next.x][next.y].Type {
	case SNAKE:
		// Eat self checks
		return errors.New("ate self")
	case FOOD:
		// A food item? Don't pop last elem (get longer) and spawn new food
		createFood()
	case EMPTY:
		// Kill final position (don't get longer)
		oldest := tailQueue.PopBack()
		board[oldest.x][oldest.y].Type = EMPTY
	}

	// Good. Plant new one on board.
	createSnake(next.x, next.y)

	return nil
}

func changeDirection(chosen Direction) {
	newest := tailQueue.At(0)
	previous := tailQueue.At(1)

	next := Cord{x: newest.x, y: newest.y}

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

	direction = chosen
}

func drawBoard() {
	wall := tcell.StyleDefault.Foreground(tcell.ColorGrey).Background(tcell.ColorGrey)
	snake := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlue)
	food := tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorRed)

	for i := 0; i < WIDTH+1; i++ {
		screen.SetContent(i, 0, tcell.RuneHLine, nil, wall)
		screen.SetContent(i, HEIGHT+1, tcell.RuneHLine, nil, wall)
	}

	for i := 0; i < HEIGHT+1; i++ {
		screen.SetContent(0, i, tcell.RuneVLine, nil, wall)
		screen.SetContent(WIDTH+1, i, tcell.RuneVLine, nil, wall)
	}

	for i := 0; i < WIDTH; i++ {
		for j := 0; j < HEIGHT; j++ {
			cell := board[i][j]
			switch cell.Type {
			case EMPTY:
			case SNAKE:
				screen.SetContent(i+1, j+1, tcell.RuneBlock, nil, snake)
			case FOOD:
				screen.SetContent(i+1, j+1, tcell.RuneDiamond, nil, food)
			}
		}
	}

}

var screen tcell.Screen

var quit = make(chan struct{})

// TODO: add score/length counters underneath walls
func main() {
	var err error

	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("Failed making screen: %v", err)
	}

	err = screen.Init()
	if err != nil {
		log.Fatalf("Failed init'ing screen: %v", err)
	}
	screen.SetStyle(tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite))

	evenChan := make(chan tcell.Event)

	go func() {
		for {
			select {
			case <-quit:
				return
			default:
			}
			ev := screen.PollEvent()
			if ev == nil {
				return
			}
			select {
			case <-quit:
				return
			case evenChan <- ev:
			case <-time.After(time.Second):
				continue
			}

		}
	}()

	ticker := time.NewTicker(100 * time.Millisecond)

	var endErr error
	initState()

LOOP:
	for {
		select {
		case <-quit:
			break LOOP
		case ev := <-evenChan:
			handleEvent(ev)
		case <-ticker.C:
			endErr = frame()
			if endErr != nil {
				break LOOP
			}
			screen.Clear()
			drawBoard()
			screen.Show()
		}
	}

	screen.Clear()
	screen.Fini()

	if endErr != nil {
		// TODO: replace this with a modal/popup with game over
		// message and option to restart with Y or N
		fmt.Printf("End: %v\n", endErr)
	} else {
		fmt.Println("Exiting")
	}
}

func handleEvent(ev tcell.Event) {

	switch ev := ev.(type) {
	case *tcell.EventInterrupt:
		close(quit)
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			close(quit)
		case tcell.KeyUp:
			changeDirection(UP)
		case tcell.KeyDown:
			changeDirection(DOWN)
		case tcell.KeyLeft:
			changeDirection(LEFT)
		case tcell.KeyRight:
			changeDirection(RIGHT)
		case tcell.KeyRune:
			if ev.Rune() == 'q' {
				close(quit)
			}
		}
	}
}
