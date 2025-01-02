package main

import (
	"fmt"
	"image/color"
	"log"

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

func main() {
	ebiten.SetWindowTitle("Snake")
	ebiten.SetWindowSize(screenWidth, screenHeight)
	if err := ebiten.RunGame(NewGame(screenWidth, screenHeight)); err != nil {
		log.Fatal(err)
	}
}

type game struct {
	Width     int
	Height    int
	Score     int
	HighScore int

	Snake Snake
	Food  Cell

	Ticks    int
	TickRate int
}

func NewGame(width, height int) *game {
	g := &game{
		Width:    width,
		Height:   height,
		Snake:    NewSnake(),
		TickRate: tickRateInitial,
	}
	g.SpawnFood()
	return g
}

func (g *game) Layout(_ int, _ int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

func (g *game) Draw(screen *ebiten.Image) {
	g.Snake.Draw(screen)
	g.Food.Draw(screen)
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
	g.HandleInput()
	if g.Ticks >= g.TickRate {
		g.Ticks = 0
		g.Snake.Move(g.Width, g.Height)
		g.CheckEat()
		g.CheckGameOver()
	}
	return nil
}

func (g *game) CheckEat() {
	if g.Snake.Head().Equals(g.Food) {
		g.Score++
		g.SpawnFood()
		g.Snake.Cells = append(g.Snake.Cells, g.Snake.Cells[len(g.Snake.Cells)-1])
		g.TickRate = max(5, g.TickRate-1)
	}
}

func (g *game) CheckGameOver() {
	head := g.Snake.Head()
	for _, cell := range g.Snake.Cells[1:] {
		if head.Equals(cell) {
			if g.Score > g.HighScore {
				g.HighScore = g.Score
			}
			g.Score = 0
			g.TickRate = tickRateInitial
			g.Snake = NewSnake()
			g.SpawnFood()
			break
		}
	}
}

func (g *game) SpawnFood() {
	food := Cell{Type: CellTypeFood}
	for {
		food.X = rand.Intn(g.Width/cellLength) * cellLength
		food.Y = rand.Intn(g.Height/cellLength) * cellLength
		available := true
		for _, cell := range g.Snake.Cells {
			if cell.Equals(food) {
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

type Snake struct {
	Cells     []Cell
	Direction Direction
}

func NewSnake() Snake {
	s := Snake{
		Cells:     make([]Cell, 3),
		Direction: DirectionRight,
	}
	for i := range s.Cells {
		s.Cells[i] = Cell{X: 100 - i*cellLength, Y: 100, Type: CellTypeSnake}
	}
	return s
}

func (s *Snake) Head() Cell {
	return s.Cells[0]
}

func (s *Snake) Move(width, height int) {
	head := s.Head()
	newHead := Cell{Type: CellTypeSnake}
	switch s.Direction {
	case DirectionUp:
		newHead.X = head.X
		newHead.Y = head.Y - cellLength
	case DirectionDown:
		newHead.X = head.X
		newHead.Y = head.Y + cellLength
	case DirectionLeft:
		newHead.X = head.X - cellLength
		newHead.Y = head.Y
	case DirectionRight:
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

	s.Cells = append([]Cell{newHead}, s.Cells[:len(s.Cells)-1]...)
}

func (s *Snake) Draw(screen *ebiten.Image) {
	for _, cell := range s.Cells {
		cell.Draw(screen)
	}
}

func (g *game) HandleInput() {
	direction := g.Snake.Direction
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp) && direction != DirectionDown:
		direction = DirectionUp
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight) && direction != DirectionLeft:
		direction = DirectionRight
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft) && direction != DirectionRight:
		direction = DirectionLeft
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown) && direction != DirectionUp:
		direction = DirectionDown
	}
	g.Snake.Direction = direction
}

type Cell struct {
	X    int
	Y    int
	Type CellType
}

type CellType int

const (
	CellTypeEmpty CellType = iota
	CellTypeSnake
	CellTypeFood
)

func (c Cell) Color() color.Color {
	switch c.Type {
	case CellTypeSnake:
		return color.RGBA{34, 139, 34, 255}
	case CellTypeFood:
		return color.White
	case CellTypeEmpty:
		fallthrough
	default:
		return color.Black
	}
}

func (c Cell) Draw(screen *ebiten.Image) {
	vector.DrawFilledRect(
		screen,
		float32(c.X),
		float32(c.Y),
		float32(cellLength),
		float32(cellLength),
		c.Color(),
		false,
	)
}

func (c Cell) Equals(other Cell) bool {
	return c.X == other.X && c.Y == other.Y
}

type Direction int

const (
	DirectionUp Direction = iota
	DirectionDown
	DirectionLeft
	DirectionRight
)
