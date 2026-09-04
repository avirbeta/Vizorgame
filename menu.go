package main

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Button — самодельная графическая кнопка с пиксельной рамкой.
// Текст храним как ключ локализации, а не готовую строку, чтобы
// кнопки мгновенно «переодевались» при смене языка.
type Button struct {
	key      string
	selfName string // у языковых кнопок — их «самоназвание»
	val      Difficulty
	x, y     float64
	w, h     float64
	hover    bool
	onTap    func()
}

func (b *Button) update() {
	mx, my := ebiten.CursorPosition()
	b.hover = float64(mx) >= b.x && float64(mx) <= b.x+b.w &&
		float64(my) >= b.y && float64(my) <= b.y+b.h
	if b.hover && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		b.onTap()
	}
}

func (b *Button) draw(screen *ebiten.Image, label string) {
	base := clrPanel
	edge := clrEdge
	if b.hover {
		base = clrPanelHi
		edge = clrAccent
	}
	// тень под кнопкой
	vector.DrawFilledRect(screen, float32(b.x+4), float32(b.y+6), float32(b.w), float32(b.h), rgb(8, 12, 20, 160), true)
	// толстая рамка
	vector.DrawFilledRect(screen, float32(b.x), float32(b.y), float32(b.w), float32(b.h), edge, true)
	// внутренний фон
	vector.DrawFilledRect(screen, float32(b.x+3), float32(b.y+3), float32(b.w-6), float32(b.h-6), base, true)
	// верхняя светлая кромка — эффект объёма
	vector.DrawFilledRect(screen, float32(b.x+3), float32(b.y+3), float32(b.w-6), 2, rgb(255, 255, 255, 45), true)
	if b.hover {
		// маленькие углы-засечки при наведении
		vector.DrawFilledRect(screen, float32(b.x-2), float32(b.y-2), 4, 4, clrAccent, true)
		vector.DrawFilledRect(screen, float32(b.x+b.w-2), float32(b.y-2), 4, 4, clrAccent, true)
		vector.DrawFilledRect(screen, float32(b.x-2), float32(b.y+b.h-2), 4, 4, clrAccent, true)
		vector.DrawFilledRect(screen, float32(b.x+b.w-2), float32(b.y+b.h-2), 4, 4, clrAccent, true)
	}

	col := clrText
	if b.hover {
		col = clrAccent
	}
	drawString(screen, label, b.x+b.w/2, b.y+b.h/2, 28, col)
}

// drawString — центральная отрисовка строки в заданной точке.
func drawString(screen *ebiten.Image, s string, cx, cy, size float64, col color.Color) {
	if s == "" {
		return
	}
	f := face(size)
	w, h := text.Measure(s, f, 0)
	op := &text.DrawOptions{}
	op.GeoM.Translate(cx-w/2, cy-h/2)
	op.ColorScale.ScaleWithColor(col)
	text.Draw(screen, s, f, op)
}

// drawStringShadow — тот же текст, но с мягкой тенью позади (заголовки).
func drawStringShadow(screen *ebiten.Image, s string, cx, cy, size float64, col color.Color) {
	drawString(screen, s, cx+2, cy+3, size, rgb(0, 0, 0, 140))
	drawString(screen, s, cx, cy, size, col)
}

// drawLang — экран 1 мастера: выбор языка.
func (g *Game) drawLang(screen *ebiten.Image) {
	drawStringShadow(screen, g.t("version"), screenW/2, 90, 46, clrAccent)
	drawString(screen, g.t("selLang"), screenW/2, 180, 26, clrDim)
	for _, b := range g.langBtns {
		b.draw(screen, b.selfName)
	}
}

// drawName — экран 2 мастера: красивое графическое поле ввода имени.
func (g *Game) drawName(screen *ebiten.Image) {
	drawStringShadow(screen, g.t("askName"), screenW/2, 210, 34, clrText)

	bx, by := float64(screenW/2-320), 300.0
	bw, bh := 640.0, 96.0
	vector.DrawFilledRect(screen, float32(bx+5), float32(by+6), float32(bw), float32(bh), rgb(8, 12, 20, 140), true)
	vector.DrawFilledRect(screen, float32(bx), float32(by), float32(bw), float32(bh), clrEdge, true)
	vector.DrawFilledRect(screen, float32(bx+3), float32(by+3), float32(bw-6), float32(bh-6), rgb(22, 34, 52, 255), true)

	face := face(42)
	label := string(g.nameBox.buf)
	w, _ := text.Measure(label, face, 0)
	tx := screenW/2 - w/2
	ty := by + (bh-42)/2
	op := &text.DrawOptions{}
	op.GeoM.Translate(tx, ty)
	op.ColorScale.ScaleWithColor(clrText)
	text.Draw(screen, label, face, op)

	// мигающий курсор
	if (g.nowTicks/35)%2 == 0 {
		vector.DrawFilledRect(screen, float32(tx+w+6), float32(by+24), 4, float32(bh-48), clrAccent, true)
	}

	drawString(screen, g.t("pressEnter"), screenW/2, 470, 22, clrDim)
	if g.warn > 0 {
		drawString(screen, g.t("emptyName"), screenW/2, 520, 24, clrDanger)
	}
}

// drawMenu — главное меню с приветствием.
func (g *Game) drawMenu(screen *ebiten.Image) {
	drawStringShadow(screen, g.t("version"), screenW/2, 90, 36, clrAccent)
	drawStringShadow(screen, fmt.Sprintf(g.t("hi"), g.name), screenW/2, 160, 38, clrText)
	for _, b := range g.menuBtns {
		b.draw(screen, g.t(b.key))
	}
	drawString(screen, "Vizor Space", screenW/2, screenH-50, 16, clrDim)
}

// drawSettings — экран настроек: язык, сложность и переназначение клавиш.
func (g *Game) drawSettings(screen *ebiten.Image) {
	drawStringShadow(screen, g.t("settings"), screenW/2, 52, 34, clrText)
	for _, b := range g.nameBtn {
		b.draw(screen, fmt.Sprintf("%s: %s", g.t("pilot"), g.name))
	}
	drawString(screen, g.t("language"), screenW/2, 92, 22, clrDim)
	for _, b := range g.setBtns {
		b.draw(screen, b.selfName)
	}

	drawString(screen, g.t("difficulty"), screenW/2, 226, 22, clrDim)
	for _, b := range g.diffBtns {
		b.draw(screen, g.t(b.key))
		if b.val == g.diff {
			vector.StrokeRect(screen, float32(b.x-3), float32(b.y-3), float32(b.w+6), float32(b.h+6), 3, clrAccent, true)
		}
	}

	drawString(screen, g.t("controls"), screenW/2, 318, 22, clrDim)
	for i, act := range actionOrder {
		g.drawCtrlRow(screen, act, i)
	}

	if g.recAct != "" {
		drawString(screen, g.t("bindNow"), screenW/2, 602, 22, clrAccent)
	} else if g.flash > 0 {
		drawString(screen, g.t(g.flashMsg), screenW/2, 602, 22, clrWin)
	}

	for _, b := range g.resetBtn {
		b.draw(screen, g.t(b.key))
	}
	for _, b := range g.backBtns {
		b.draw(screen, g.t(b.key))
	}
}

// drawRecords — таблица лучших пилотов из highscores.dat.
func (g *Game) drawRecords(screen *ebiten.Image) {
	drawStringShadow(screen, g.t("records"), screenW/2, 80, 40, clrAccent)

	if len(g.recs) == 0 {
		drawString(screen, g.t("noRecords"), screenW/2, 260, 26, clrDim)
	} else {
		y := 200.0
		for i, r := range g.recs {
			var col color.Color = clrWin
			switch i {
			case 1:
				col = rgb(240, 226, 170, 255)
			case 2:
				col = rgb(210, 170, 120, 255)
			default:
				col = clrText
			}
			line := fmt.Sprintf("%d.  %s", i+1, r.Name)
			drawString(screen, line, screenW/2-140, y, 30, col)
			drawString(screen, fmt.Sprintf("%d", r.Score), screenW/2+180, y, 30, col)
			y += 64
		}
	}

	for _, b := range g.backBtns {
		b.draw(screen, g.t(b.key))
	}
}

// drawHelp — экран инструкции: цель игры и текущее управление по центру.
func (g *Game) drawHelp(screen *ebiten.Image) {
	drawStringShadow(screen, g.t("help"), screenW/2, 60, 38, clrText)

	drawString(screen, g.t("goal")+":", screenW/2, 150, 26, clrAccent)
	drawString(screen, g.t("goalText"), screenW/2, 190, 22, clrText)

	drawString(screen, g.t("controls")+":", screenW/2, 270, 26, clrAccent)
	y := 320.0
	for _, act := range actionOrder {
		line := fmt.Sprintf("%s  —  %s", g.t(langKeyForAction(act)), g.keys.names(act))
		drawString(screen, line, screenW/2, y, 26, clrText)
		y += 62
	}
	drawString(screen, g.t("pauseHint"), screenW/2, y+20, 20, clrDim)

	for _, b := range g.backBtns {
		b.draw(screen, g.t(b.key))
	}
}

// геометрия рядов управления и плашек-клавиш в настройках:
// вынес в константы, потому что при каждой переделке всё съезжает
const (
	ctrlRowTop0 = 354.0
	ctrlRowH    = 50.0
	ctrlRowGap  = 10.0
	chipX0      = 790.0
	chipW       = 150.0
	chipH       = 36.0
	chipGap     = 18.0
)

func ctrlRowTop(i int) float64 {
	return ctrlRowTop0 + float64(i)*(ctrlRowH+ctrlRowGap)
}

func chipX(j int) float64 {
	return chipX0 + float64(j)*(chipW+chipGap)
}

// drawCtrlRow — строка «действие → клавиша» в настройках.
func (g *Game) drawCtrlRow(screen *ebiten.Image, act action, i int) {
	top := ctrlRowTop(i)

	panel := rgb(22, 32, 48, 255)
	if g.recAct == act {
		panel = rgb(70, 66, 40, 255)
	}
	vector.DrawFilledRect(screen, 320, float32(top), 810, float32(ctrlRowH), panel, true)
	vector.StrokeRect(screen, 320, float32(top), 810, float32(ctrlRowH), 2, clrEdge, true)

	label := g.t(langKeyForAction(act))
	drawString(screen, label, 360, top+ctrlRowH/2, 24, clrText)

	keys := g.keys[act]
	for j := 0; j < 2; j++ {
		x := chipX(j)
		if j < len(keys) {
			g.drawKeyChip(screen, x, top, keyLabel(keys[j]), g.recAct == act && g.recSlot == j)
		} else {
			g.drawKeyChip(screen, x, top, "+", false)
		}
	}
}

// drawKeyChip — небольшая плашка с названием клавиши.
func (g *Game) drawKeyChip(screen *ebiten.Image, x, rowTop float64, label string, active bool) {
	mx, my := ebiten.CursorPosition()
	y := rowTop + (ctrlRowH-chipH)/2
	hover := float64(mx) >= x && float64(mx) <= x+chipW && float64(my) >= y && float64(my) <= y+chipH

	var col color.Color = clrPanelHi
	border := clrEdge
	txt := clrText
	if active {
		col = rgb(96, 84, 40, 255)
		border = clrAccent
		txt = clrAccent
	} else if hover {
		border = clrAccent
		col = clrPanel
	}

	vector.DrawFilledRect(screen, float32(x), float32(y), float32(chipW), float32(chipH), col, true)
	vector.StrokeRect(screen, float32(x), float32(y), float32(chipW), float32(chipH), 2, border, true)
	drawString(screen, label, x+chipW/2, y+chipH/2, 22, txt)
}

// trackChips — ловим клики по «чипам» клавиш и включаем режим ожидания
// новой клавиши для выбранного действия.
func (g *Game) trackChips() {
	if !inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		return
	}
	mx, my := ebiten.CursorPosition()
	for i, act := range actionOrder {
		top := ctrlRowTop(i)
		for j := 0; j < 2; j++ {
			x := chipX(j)
			y := top + (ctrlRowH-chipH)/2
			if float64(mx) >= x && float64(mx) <= x+chipW && float64(my) >= y && float64(my) <= y+chipH {
				g.recAct = act
				g.recSlot = j
				return
			}
		}
	}
}

// меню-фоновые звёздочки, чтобы экраны не выглядели мёртвой пустотой
func (g *Game) drawAmbient(screen *ebiten.Image) {
	for _, s := range g.amb {
		vector.DrawFilledCircle(screen, float32(s.x), float32(s.y), float32(s.r), rgb(120, 150, 190, s.a), true)
	}
}
