package main

import (
	"encoding/json"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// action — смысловое имя игрового действия. Клавиши для него
// игрок может менять в настройках.
type action string

const (
	actLeft  action = "left"
	actRight action = "right"
	actShoot action = "shoot"
	actBack  action = "back"
)

// actionOrder — порядок, в котором действия показываются в настройках.
var actionOrder = []action{actLeft, actRight, actShoot, actBack}

// langKeyForAction переводит действие в ключ локализации для его названия.
func langKeyForAction(a action) string {
	switch a {
	case actLeft:
		return "kLeft"
	case actRight:
		return "kRight"
	case actShoot:
		return "kShoot"
	case actBack:
		return "kBack"
	}
	return "?"
}

// KeyMap — привязка действия к одной-двум клавишам (вторая — запасная,
// та же польза, что у WASD рядом со стрелками).
type KeyMap map[action][]ebiten.Key

func defaultKeyMap() KeyMap {
	return KeyMap{
		actLeft:  {ebiten.KeyLeft, ebiten.KeyA},
		actRight: {ebiten.KeyRight, ebiten.KeyD},
		actShoot: {ebiten.KeySpace},
		actBack:  {ebiten.KeyEscape},
	}
}

// held — зажата ли хотя бы одна клавиша действия.
func (km KeyMap) held(a action) bool {
	for _, k := range km[a] {
		if ebiten.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

// justDown — было ли действие только что нажато.
func (km KeyMap) justDown(a action) bool {
	for _, k := range km[a] {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

// names — красивое перечисление клавиш через слэш для справки.
func (km KeyMap) names(a action) string {
	parts := make([]string, 0, len(km[a]))
	for _, k := range km[a] {
		parts = append(parts, keyLabel(k))
	}
	return strings.Join(parts, " / ")
}

// keyLabel — короткое человекочитаемое имя клавиши.
// ebiten.Key.String() возвращает что-то вроде "ArrowLeft" или "Digit5",
// мы снимаем эти служебные приставки и оставляем суть.
func keyLabel(k ebiten.Key) string {
	s := k.String()
	s = strings.TrimPrefix(s, "Key")
	s = strings.TrimPrefix(s, "Digit")
	s = strings.TrimPrefix(s, "Arrow")
	return s
}

// bindableKeys — перечень клавиш, которые можно поставить на действие.
var bindableKeys = []ebiten.Key{
	ebiten.KeyA, ebiten.KeyB, ebiten.KeyC, ebiten.KeyD, ebiten.KeyE,
	ebiten.KeyF, ebiten.KeyG, ebiten.KeyH, ebiten.KeyI, ebiten.KeyJ,
	ebiten.KeyK, ebiten.KeyL, ebiten.KeyM, ebiten.KeyN, ebiten.KeyO,
	ebiten.KeyP, ebiten.KeyQ, ebiten.KeyR, ebiten.KeyS, ebiten.KeyT,
	ebiten.KeyU, ebiten.KeyV, ebiten.KeyW, ebiten.KeyX, ebiten.KeyY,
	ebiten.KeyZ,
	ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4,
	ebiten.KeyDigit5, ebiten.KeyDigit6, ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9,
	ebiten.KeySpace, ebiten.KeyEnter, ebiten.KeyEscape, ebiten.KeyTab,
	ebiten.KeyArrowLeft, ebiten.KeyArrowRight, ebiten.KeyArrowUp, ebiten.KeyArrowDown,
	ebiten.KeyShift, ebiten.KeyControl, ebiten.KeyAlt,
}

// grabKey ищет любую только что нажатую клавишу из разрешённого списка.
// Так работает режим переназначения: игрок кликнул по «чипу», а мы
// слушаем, какую клавишу он нажмёт следующей.
func grabKey() (ebiten.Key, bool) {
	for _, k := range bindableKeys {
		if inpututil.IsKeyJustPressed(k) {
			return k, true
		}
	}
	return 0, false
}

// rebindKey ставит клавишу k на позицию slot действия act. Прежде всего
// клавишу убираем у всех остальных действий — двойников не держим.
func (g *Game) rebindKey(act action, slot int, k ebiten.Key) {
	for a := range g.keys {
		kept := g.keys[a][:0]
		for _, old := range g.keys[a] {
			if old != k {
				kept = append(kept, old)
			}
		}
		g.keys[a] = kept
	}

	ks := g.keys[act]
	for len(ks) <= slot {
		ks = append(ks, 0)
	}
	ks[slot] = k
	if len(ks) > 2 {
		ks = ks[:2]
	}
	g.keys[act] = ks
}

// encodeKeys превращает текущую раскладку в строку JSON для save.dat.
func encodeKeys(km KeyMap) string {
	b, err := json.Marshal(km)
	if err != nil {
		return ""
	}
	return string(b)
}

// decodeKeys читает раскладку из JSON; пустую/битую строку подменяем
// стандартной раскладкой.
func decodeKeys(s string) KeyMap {
	km := defaultKeyMap()
	if s == "" {
		return km
	}
	var raw map[action][]ebiten.Key
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		return km
	}
	for _, a := range actionOrder {
		if ks, ok := raw[a]; ok && len(ks) > 0 {
			km[a] = ks
		}
	}
	return km
}
