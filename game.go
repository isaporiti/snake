package game

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/exp/rand"
	"golang.org/x/image/font/basicfont"
)

const (
	screenWidth     = 200
	screenHeight    = 300
	cellLength      = 10
	tickRateInitial = 20
)

func Run() error {
	ebiten.SetWindowTitle("Snake")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	return ebiten.RunGame(newGame(screenWidth, screenHeight))
}

type game struct {
	Width     int
	Height    int
	Score     int
	HighScore int

	Snake snake
	Food  cell

	Ticks    int
	TickRate int
}

func newGame(width, height int) *game {
	g := &game{
		Width:    width,
		Height:   height,
		Snake:    newSnake(),
		TickRate: tickRateInitial,
	}
	g.spawnFood()
	return g
}

func (g *game) Layout(_ int, _ int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

func (g *game) Draw(screen *ebiten.Image) {
	g.Snake.Draw(screen)
	g.Food.draw(screen)
	text.Draw(
		screen,
		fmt.Sprintf("Score: %d", g.Score),
		basicfont.Face7x13, 10, 20, color.White,
	)
	text.Draw(
		screen,
		fmt.Sprintf("High Score: %d", g.HighScore),
		basicfont.Face7x13, 10, 40, color.White,
	)
}

func (g *game) Update() error {
	g.Ticks++
	g.handleInput()
	if g.Ticks >= g.TickRate {
		g.Ticks = 0
		g.Snake.move(g.Width, g.Height)
		g.checkEat()
		g.checkGameOver()
	}
	return nil
}

func (g *game) checkEat() {
	if g.Snake.head().equals(g.Food) {
		g.Score++
		g.spawnFood()
		g.Snake.Cells = append(g.Snake.Cells, g.Snake.Cells[len(g.Snake.Cells)-1])
		g.TickRate = max(5, g.TickRate-1)
	}
}

func (g *game) checkGameOver() {
	head := g.Snake.head()
	for _, cell := range g.Snake.Cells[1:] {
		if head.equals(cell) {
			if g.Score > g.HighScore {
				g.HighScore = g.Score
			}
			g.Score = 0
			g.TickRate = tickRateInitial
			g.Snake = newSnake()
			g.spawnFood()
			break
		}
	}
}

func (g *game) spawnFood() {
	food := cell{Type: cellTypeFood}
	for {
		food.X = rand.Intn(g.Width/cellLength) * cellLength
		food.Y = rand.Intn(g.Height/cellLength) * cellLength
		available := true
		for _, cell := range g.Snake.Cells {
			if cell.equals(food) {
				available = false
				break
			}
		}
		if available {
			break
		}
	}
	g.Food = food
}

type snake struct {
	Cells     []cell
	Direction direction
}

func newSnake() snake {
	s := snake{
		Cells:     make([]cell, 3),
		Direction: directionRight,
	}
	for i := range s.Cells {
		s.Cells[i] = cell{X: 100 - i*cellLength, Y: 100, Type: cellTypeSnake}
	}
	return s
}

func (s *snake) head() cell {
	return s.Cells[0]
}

func (s *snake) move(width, height int) {
	head := s.head()
	newHead := cell{Type: cellTypeSnake}
	switch s.Direction {
	case directionUp:
		newHead.X = head.X
		newHead.Y = head.Y - cellLength
	case directionDown:
		newHead.X = head.X
		newHead.Y = head.Y + cellLength
	case directionLeft:
		newHead.X = head.X - cellLength
		newHead.Y = head.Y
	case directionRight:
		newHead.X = head.X + cellLength
		newHead.Y = head.Y
	}

	if newHead.X < 0 {
		newHead.X = width - cellLength
	} else if newHead.X >= width {
		newHead.X = 0
	}
	if newHead.Y < 0 {
		newHead.Y = height - cellLength
	} else if newHead.Y >= height {
		newHead.Y = 0
	}

	s.Cells = append([]cell{newHead}, s.Cells[:len(s.Cells)-1]...)
}

func (s *snake) Draw(screen *ebiten.Image) {
	for _, cell := range s.Cells {
		cell.draw(screen)
	}
}

func (g *game) handleInput() {
	direction := g.Snake.Direction
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp) && direction != directionDown:
		direction = directionUp
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight) && direction != directionLeft:
		direction = directionRight
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft) && direction != directionRight:
		direction = directionLeft
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown) && direction != directionUp:
		direction = directionDown
	}
	g.Snake.Direction = direction
}

type cell struct {
	X    int
	Y    int
	Type cellType
}

type cellType int

const (
	cellTypeEmpty cellType = iota
	cellTypeSnake
	cellTypeFood
)

func (c cell) color() color.Color {
	switch c.Type {
	case cellTypeSnake:
		return color.RGBA{34, 139, 34, 255}
	case cellTypeFood:
		return color.White
	case cellTypeEmpty:
		fallthrough
	default:
		return color.Black
	}
}

func (c cell) draw(screen *ebiten.Image) {
	vector.DrawFilledRect(
		screen,
		float32(c.X),
		float32(c.Y),
		float32(cellLength),
		float32(cellLength),
		c.color(),
		false,
	)
}

func (c cell) equals(other cell) bool {
	return c.X == other.X && c.Y == other.Y
}

type direction int

const (
	directionUp direction = iota
	directionDown
	directionLeft
	directionRight
)
