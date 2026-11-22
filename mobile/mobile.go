// Package vizorgamemobile — точка входа для Android/iOS-обёртки.
// ebitenmobile bind собирает её в AAR, внутри которого живёт EbitenView.
package vizorgamemobile

import (
	"github.com/hajimehoshi/ebiten/v2/mobile"

	"github.com/avirbeta/Vizorgame/game"
)

func init() {
	mobile.SetGame(game.NewGame())
}

// Dummy нужен, чтобы gomobile не выбросил пакет без экспортируемых функций.
func Dummy() {}
