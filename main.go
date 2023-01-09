package main

import (
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell"
)

type UI struct {
	screen tcell.Screen
	snake  *Snake
	paused bool
	quit   chan struct{}
}

const (
	WIDTH  = 30
	HEIGHT = 30
)

func (g *UI) drawBoard() {
	wall := tcell.StyleDefault.Foreground(tcell.ColorGrey).Background(tcell.ColorGrey)
	snake := tcell.StyleDefault.Foreground(tcell.ColorBlue).Background(tcell.ColorBlue)
	food := tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorRed)

	for i := 0; i < WIDTH+1; i++ {
		g.screen.SetContent(i, 0, tcell.RuneHLine, nil, wall)
		g.screen.SetContent(i, HEIGHT+1, tcell.RuneHLine, nil, wall)
	}

	for i := 0; i < HEIGHT+1; i++ {
		g.screen.SetContent(0, i, tcell.RuneVLine, nil, wall)
		g.screen.SetContent(WIDTH+1, i, tcell.RuneVLine, nil, wall)
	}

	for cord, cell := range g.snake.board {
		switch cell.Type {
		case SNAKE:
			g.screen.SetContent(cord.x+1, cord.y+1, tcell.RuneBlock, nil, snake)
		case FOOD:
			g.screen.SetContent(cord.x+1, cord.y+1, tcell.RuneDiamond, nil, food)
		}
	}
}

// TODO: add score/length counters underneath walls
func main() {

	screen, err := tcell.NewTerminfoScreen()
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

	ui := &UI{
		screen: screen,
		snake:  &Snake{},
	}
	ui.Run()
}

func (g *UI) Run() {
	evenChan := make(chan tcell.Event)

	go func() {
		for {
			select {
			case <-g.quit:
				return
			default:
			}
			ev := g.screen.PollEvent()
			if ev == nil {
				return
			}
			select {
			case <-g.quit:
				return
			case evenChan <- ev:
			case <-time.After(time.Second):
				continue
			}

		}
	}()

	ticker := time.NewTicker(100 * time.Millisecond)

	var endErr error
	g.snake.Reset()

LOOP:
	for {
		select {
		case <-g.quit:
			break LOOP
		case ev := <-evenChan:
			g.handleEvent(ev)
		case <-ticker.C:
			if !g.paused {
				endErr = g.snake.UpdateState()
				if endErr != nil {
					break LOOP
				}
			}
			g.screen.Clear()
			g.drawBoard()
			g.screen.Show()
		}
	}

	g.screen.Clear()
	g.screen.Fini()

	if endErr != nil {
		// TODO: replace this with a modal/popup with game over
		// message and option to restart with Y or N
		fmt.Printf("End: %v\n", endErr)
	} else {
		fmt.Println("Exiting")
	}

}

func (g *UI) handleEvent(ev tcell.Event) {

	switch ev := ev.(type) {
	case *tcell.EventInterrupt:
		close(g.quit)
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			close(g.quit)
		case tcell.KeyUp:
			g.snake.ChangeDirection(UP)
		case tcell.KeyDown:
			g.snake.ChangeDirection(DOWN)
		case tcell.KeyLeft:
			g.snake.ChangeDirection(LEFT)
		case tcell.KeyRight:
			g.snake.ChangeDirection(RIGHT)
		case tcell.KeyRune:
			switch ev.Rune() {
			case 'q':
				close(g.quit)
			case 'p':
				g.paused = !g.paused
			}
		}
	}
}
