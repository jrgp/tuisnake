package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gdamore/tcell"
)

type UI struct {
	screen   tcell.Screen
	snake    *Snake
	paused   bool
	gameover bool
	cancel   context.CancelFunc
}

const (
	WIDTH  = 30
	HEIGHT = 30
)

func (g *UI) drawBoard() {
	wall := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGreen)
	snake := tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorBlue)
	food := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorRed)

	for i := 0; i < WIDTH; i++ {
		g.screen.SetContent(i, 0, tcell.RuneHLine, nil, wall)
		g.screen.SetContent(i, HEIGHT, tcell.RuneHLine, nil, wall)
	}

	for i := 0; i < HEIGHT; i++ {
		g.screen.SetContent(0, i, tcell.RuneVLine, nil, wall)
		g.screen.SetContent(WIDTH, i, tcell.RuneVLine, nil, wall)
	}
	g.screen.SetContent(0, HEIGHT, tcell.RuneLLCorner, nil, wall)
	g.screen.SetContent(WIDTH, HEIGHT, tcell.RuneLRCorner, nil, wall)

	g.screen.SetContent(WIDTH, 0, tcell.RuneURCorner, nil, wall)
	g.screen.SetContent(0, 0, tcell.RuneULCorner, nil, wall)

	for cord, cell := range g.snake.board {
		switch cell.Type {
		case SNAKE:
			g.screen.SetContent(cord.x+1, cord.y+1, tcell.RuneBlock, nil, snake)
		case FOOD:
			g.screen.SetContent(cord.x+1, cord.y+1, tcell.RuneDiamond, nil, food)
		}
	}

	subtitle := fmt.Sprintf("Size: %v", g.snake.tailQueue.Len())

	if g.gameover {
		subtitle = "GAME OVER. 'R' to replay"
	} else {
		if g.paused {
			subtitle += " (PAUSED)"
		}
	}

	for i, char := range subtitle {
		g.screen.SetContent(i+1, HEIGHT+1, char, nil, tcell.StyleDefault)
	}
}

func (g *UI) Run() {
	evenChan := make(chan tcell.Event)
	ctx, cancel := context.WithCancel(context.Background())
	g.cancel = cancel

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			ev := g.screen.PollEvent()
			if ev == nil {
				return
			}
			select {
			case <-ctx.Done():
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
		case <-ctx.Done():
			break LOOP
		case ev := <-evenChan:
			g.handleEvent(ev)
		case <-ticker.C:
			if !g.paused && !g.gameover {
				endErr = g.snake.UpdateState()
				if endErr != nil {
					g.gameover = true
				}
			}
			g.screen.Clear()
			g.drawBoard()
			g.screen.Show()
		}
	}

	g.screen.Clear()
	g.screen.Fini()
}

func (g *UI) handleEvent(ev tcell.Event) {

	switch ev := ev.(type) {
	case *tcell.EventInterrupt:
		g.cancel()
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			g.cancel()
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
				g.cancel()
			case 'p':
				g.paused = !g.paused
			case 'r':
				g.snake.Reset()
				g.gameover = false
				g.paused = false
			}
		}
	}
}

func main() {

	screen, err := tcell.NewScreen()
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
