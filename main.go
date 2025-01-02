package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

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
	tickRateInitial = 15
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

	Touches       []ebiten.TouchID
	TouchId       ebiten.TouchID
	TouchState    TouchState
	TouchInitPosX int
	TouchInitPosY int
	TouchLastPosX int
	TouchLastPosY int
}

type TouchState uint8

const (
	TouchStateNone = iota
	TouchStatePressing
	TouchStateSettled
	TouchStateInvalid
)

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
	case ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		direction = DirectionUp
	case ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		direction = DirectionRight
	case ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		direction = DirectionLeft
	case ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		direction = DirectionDown
	}

	g.Touches = ebiten.AppendTouchIDs(g.Touches[:0])
	switch g.TouchState {
	case TouchStateNone:
		if len(g.Touches) == 1 {
			g.TouchId = g.Touches[0]
			x, y := ebiten.TouchPosition(g.TouchId)
			g.TouchInitPosX = x
			g.TouchInitPosY = y
			g.TouchLastPosX = x
			g.TouchLastPosY = y
			g.TouchState = TouchStatePressing
		}
	case TouchStatePressing:
		if len(g.Touches) >= 2 {
			break
		}
		if len(g.Touches) == 1 {
			if g.Touches[0] != g.TouchId {
				g.TouchState = TouchStateInvalid
			} else {
				x, y := ebiten.TouchPosition(g.Touches[0])
				g.TouchLastPosX = x
				g.TouchLastPosY = y
			}
			break
		}
		if len(g.Touches) == 0 {
			direction = g.vecToDir(
				float64(g.TouchLastPosX-g.TouchInitPosX),
				float64(g.TouchLastPosY-g.TouchInitPosY),
			)
			if direction == DirectionNone {
				g.TouchState = TouchStateNone
				break
			}
			g.TouchState = TouchStateSettled
		}
	case TouchStateSettled:
		g.TouchState = TouchStateNone
	case TouchStateInvalid:
		if len(g.Touches) == 0 {
			g.TouchState = TouchStateNone
		}
	}

	switch {
	case direction == DirectionUp && g.Snake.Direction != DirectionDown:
		fallthrough
	case direction == DirectionDown && g.Snake.Direction != DirectionUp:
		fallthrough
	case direction == DirectionLeft && g.Snake.Direction != DirectionRight:
		fallthrough
	case direction == DirectionRight && g.Snake.Direction != DirectionLeft:
		g.Snake.Direction = direction
	}
}

func (g *game) vecToDir(x, y float64) Direction {
	if math.Abs(x) < 4 && math.Abs(y) < 4 {
		return DirectionNone
	}
	if math.Abs(x) < math.Abs(y) {
		if y < 0 {
			return DirectionUp
		}
		return DirectionDown
	}
	if x < 0 {
		return DirectionLeft
	}
	return DirectionRight
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
	DirectionNone Direction = iota
	DirectionUp
	DirectionDown
	DirectionLeft
	DirectionRight
)
