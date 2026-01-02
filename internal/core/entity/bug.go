package entity

import "github.com/gdamore/tcell/v2"

type Bug struct {
	BaseEnemy
	isArmored bool
}

func NewBug(x, y int, armored bool) *Bug {
	hp := 1
	char := '👾'
	style := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)

	if armored {
		hp = 10
		char = '🛡'
		style = tcell.StyleDefault.Foreground(tcell.ColorBlue).Bold(true)
	}

	return &Bug{
		BaseEnemy: BaseEnemy{
			X:     float64(x),
			Y:     float64(y),
			VelY:  2.5,
			Char:  char,
			Style: style,
			HP:    hp, MaxHP: hp,
		},
		isArmored: armored,
	}
}

func (e *Bug) Update(ctx Context) {
	e.X += e.VelX * ctx.DT
	e.Y += e.VelY * ctx.DT
}

func (e *Bug) IsArmored() bool {
	return e.isArmored
}
