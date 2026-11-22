package game

import (
	"image"
	"image/color"
	"math/rand"
	"os"
	"runtime"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenW = 1280
	screenH = 720
)

// Экранные кнопки для телефонов/планшетов.
const (
	touchLeftX   = 120
	touchLeftY   = 596
	touchRightX  = 290
	touchRightY  = 596
	touchButtonR = 68
	touchFireX   = 1140
	touchFireY   = 596
	touchFireR   = 84
)

// screenID — состояние верхнего уровня. Игровая аркада живёт внутри scrGame.
type screenID int

const (
	scrLang screenID = iota
	scrName
	scrMenu
	scrSettings
	scrRecords
	scrHelp
	scrGame
)

// Палитра интерфейса.
var (
	clrText    = rgb(235, 240, 250, 255)
	clrDim     = rgb(150, 165, 190, 255)
	clrAccent  = rgb(120, 200, 255, 255)
	clrDanger  = rgb(255, 110, 110, 255)
	clrWin     = rgb(150, 255, 180, 255)
	clrPanel   = rgb(18, 26, 40, 255)
	clrPanelHi = rgb(28, 40, 60, 255)
	clrEdge    = rgb(70, 95, 130, 255)
	clrShip    = rgb(150, 220, 255, 255)
	clrBullet  = rgb(255, 235, 150, 255)
)

func rgb(r, g, b, a uint8) color.RGBA {
	return color.RGBA{R: r, G: g, B: b, A: a}
}

// Game — корневой объект ebiten: хранит экран, профиль и настройки.
type Game struct {
	scr       screenID
	afterName screenID // куда вернуться после ввода имени

	langID LangID
	name   string
	keys   KeyMap
	diff   Difficulty
	cfg    config

	savePath string
	best     int
	recs     []record

	sfx *sfxChip
	arc *Arcade

	nowTicks int
	warn     int
	flash    int
	flashMsg string

	recAct  action
	recSlot int

	langBtns []*Button
	menuBtns []*Button
	nameBtn  []*Button
	setBtns  []*Button
	diffBtns []*Button
	resetBtn []*Button
	backBtns []*Button
	nameBox  NameField
	amb      []*spark
	quit     bool

	touchMode     bool
	touchLeft     bool
	touchRight    bool
	touchFire     bool
	touchFireJust bool
	touchJust     bool
	touchX        int
	touchY        int
	prevTouches   map[ebiten.TouchID]bool
}

func NewGame() *Game {
	cfg := loadConfig()
	g := &Game{
		cfg:       cfg,
		langID:    LangEN,
		keys:      defaultKeyMap(),
		diff:      DiffNormal,
		afterName: scrMenu,
	}
	g.savePath = resolveSavePath(cfg)
	if cfg.hasLang {
		g.langID = cfg.lang
	}

	lang, name, keys, diff, ok := readSave(g.savePath)
	if ok {
		g.langID = lang
		g.keys = keys
		g.diff = diff
		g.name = name
		g.nameBox.buf = []byte(name)
		if name == "" {
			g.scr = scrName
		} else {
			g.scr = scrMenu
		}
	} else if cfg.hasLang {
		g.scr = scrName
	} else {
		g.scr = scrLang
	}

	g.sfx = nil
	if audioAvailable() {
		g.sfx = newSfx(cfg.volume)
	}

	g.touchMode = runtime.GOOS == "android" || os.Getenv("VIZOR_TOUCH") == "1"
	g.prevTouches = map[ebiten.TouchID]bool{}
	g.recs = loadRecords(g.savePath)
	g.best = bestScore(g.recs)

	for i := 0; i < 70; i++ {
		g.amb = append(g.amb, &spark{
			x: rand.Float64() * screenW,
			y: rand.Float64() * screenH,
			r: 0.6 + rand.Float64()*1.8,
			a: uint8(30 + rand.Intn(70)),
		})
	}

	g.buildButtons()
	return g
}

func (g *Game) buildButtons() {
	g.langBtns = nil
	for i, item := range langOrder {
		lang := item.ID
		g.langBtns = append(g.langBtns, &Button{
			selfName: item.Self,
			x:        screenW/2 - 210,
			y:        230 + float64(i)*74,
			w:        420,
			h:        58,
			onTap: func() {
				g.setLanguage(lang)
				g.change(scrName)
			},
		})
	}

	menu := []struct {
		key    string
		target screenID
	}{
		{"play", scrGame},
		{"settings", scrSettings},
		{"records", scrRecords},
		{"help", scrHelp},
	}
	g.menuBtns = nil
	for i, item := range menu {
		target := item.target
		g.menuBtns = append(g.menuBtns, &Button{
			key: item.key,
			x:   screenW/2 - 210,
			y:   240 + float64(i)*68,
			w:   420,
			h:   54,
			onTap: func() {
				if target == scrGame {
					g.startGame()
					return
				}
				g.change(target)
			},
		})
	}
	g.menuBtns = append(g.menuBtns, &Button{
		key: "exit",
		x:   screenW/2 - 210,
		y:   240 + float64(len(menu))*68,
		w:   420,
		h:   54,
		onTap: func() {
			g.saveProfile()
			g.quit = true
		},
	})

	g.nameBtn = []*Button{{
		x: screenW/2 - 280,
		y: 62,
		w: 560,
		h: 40,
		onTap: func() {
			g.afterName = scrSettings
			g.nameBox.buf = []byte(g.name)
			g.change(scrName)
		},
	}}

	g.setBtns = nil
	for i, item := range langOrder {
		lang := item.ID
		g.setBtns = append(g.setBtns, &Button{
			selfName: item.Self,
			x:        screenW/2 - 360 + float64(i)*182,
			y:        134,
			w:        168,
			h:        38,
			onTap: func() {
				g.setLanguage(lang)
			},
		})
	}

	diffs := []struct {
		key string
		val Difficulty
	}{
		{"dEasy", DiffEasy},
		{"dNormal", DiffNormal},
		{"dHard", DiffHard},
	}
	g.diffBtns = nil
	for i, item := range diffs {
		val := item.val
		g.diffBtns = append(g.diffBtns, &Button{
			key: item.key,
			val: val,
			x:   screenW/2 - 245 + float64(i)*170,
			y:   206,
			w:   150,
			h:   40,
			onTap: func() {
				g.diff = val
				g.saveProfile()
				g.flashMsg = "saved"
				g.flash = 100
			},
		})
	}

	g.resetBtn = []*Button{{
		key: "resetKeys",
		x:   screenW/2 - 300,
		y:   606,
		w:   280,
		h:   44,
		onTap: func() {
			g.keys = defaultKeyMap()
			g.saveProfile()
			g.flashMsg = "resetDone"
			g.flash = 110
		},
	}}

	g.backBtns = []*Button{{
		key: "back",
		x:   screenW/2 + 20,
		y:   606,
		w:   280,
		h:   44,
		onTap: func() {
			g.change(scrMenu)
		},
	}}
}

func (g *Game) activeButtons() []*Button {
	switch g.scr {
	case scrLang:
		return g.langBtns
	case scrMenu:
		return g.menuBtns
	case scrSettings:
		buttons := make([]*Button, 0, 12)
		buttons = append(buttons, g.nameBtn...)
		buttons = append(buttons, g.setBtns...)
		buttons = append(buttons, g.diffBtns...)
		buttons = append(buttons, g.resetBtn...)
		buttons = append(buttons, g.backBtns...)
		return buttons
	case scrRecords, scrHelp:
		return g.backBtns
	default:
		return nil
	}
}

// readTouch обновляет экранные кнопки для игры на телефоне.
func (g *Game) readTouch() {
	current := make(map[ebiten.TouchID]bool, 4)
	g.touchLeft = false
	g.touchRight = false
	g.touchFire = false
	g.touchFireJust = false
	g.touchJust = false
	g.touchX, g.touchY = -1, -1

	for _, id := range ebiten.AppendTouchIDs(nil) {
		current[id] = true
		x, y := ebiten.TouchPosition(id)
		isNew := !g.prevTouches[id]

		if isNew {
			g.touchJust = true
			if g.touchX < 0 {
				g.touchX, g.touchY = x, y
			}
		}

		switch {
		case touchHit(x, y, touchLeftX, touchLeftY, touchButtonR):
			g.touchLeft = true
		case touchHit(x, y, touchRightX, touchRightY, touchButtonR):
			g.touchRight = true
		case touchHit(x, y, touchFireX, touchFireY, touchFireR):
			g.touchFire = true
			if isNew {
				g.touchFireJust = true
			}
		}
	}

	g.prevTouches = current
}

func touchHit(x, y, cx, cy, radius int) bool {
	dx := x - cx
	dy := y - cy
	return dx*dx+dy*dy <= radius*radius
}

func (g *Game) drawTouchControls(screen *ebiten.Image) {
	if !g.touchMode || g.scr != scrGame {
		return
	}
	drawTouchButton(screen, touchLeftX, touchLeftY, touchButtonR, "◀", g.touchLeft)
	drawTouchButton(screen, touchRightX, touchRightY, touchButtonR, "▶", g.touchRight)
	drawTouchButton(screen, touchFireX, touchFireY, touchFireR, "●", g.touchFire)
}

func drawTouchButton(screen *ebiten.Image, cx, cy, radius int, label string, active bool) {
	fill := color.RGBA{R: 30, G: 45, B: 65, A: 150}
	if active {
		fill = color.RGBA{R: 80, G: 140, B: 190, A: 200}
	}
	vector.DrawFilledCircle(screen, float32(cx), float32(cy), float32(radius), fill, true)
	vector.StrokeCircle(screen, float32(cx), float32(cy), float32(radius), 3, clrAccent, true)
	drawString(screen, label, float64(cx), float64(cy), 34, clrText)
}

func (g *Game) setLanguage(lang LangID) {
	g.langID = lang
	g.cfg.lang = lang
	g.cfg.hasLang = true
	saveConfig(g.cfg)
	if g.name != "" {
		writeSave(g.savePath, g.langID, g.name, g.keys, g.diff)
	}
}

func (g *Game) saveProfile() {
	g.cfg.lang = g.langID
	g.cfg.hasLang = true
	saveConfig(g.cfg)
	if g.name == "" {
		return
	}
	writeSave(g.savePath, g.langID, g.name, g.keys, g.diff)
}

func (g *Game) submitName() {
	name := strings.TrimSpace(g.nameBox.name())
	if name == "" {
		g.warn = 90
		return
	}
	g.name = name
	g.nameBox.buf = []byte(name)
	g.saveProfile()
	g.flashMsg = "saved"
	g.flash = 90
	g.change(g.afterName)
}

func (g *Game) startGame() {
	g.arc = newArcade()
	g.change(scrGame)
}

func (g *Game) change(next screenID) {
	g.scr = next
	g.recAct = ""
	g.recSlot = 0
	g.warn = 0
	g.flash = 0
	if next == scrRecords {
		g.recs = loadRecords(g.savePath)
		g.best = bestScore(g.recs)
	}
	if next == scrGame && g.arc == nil {
		g.arc = newArcade()
	}
}

func (g *Game) Update() error {
	g.nowTicks++
	if g.flash > 0 {
		g.flash--
	}
	if g.warn > 0 {
		g.warn--
	}
	for _, s := range g.amb {
		s.y += 0.18
		if s.y > screenH {
			s.y = -2
			s.x = rand.Float64() * screenW
		}
	}

	if g.touchMode {
		g.readTouch()
	}

	buttons := g.activeButtons()
	for _, button := range buttons {
		button.update()
	}
	if g.touchMode && g.touchJust {
		for _, button := range buttons {
			button.touchAt(g.touchX, g.touchY)
		}
	}

	if g.quit {
		return ebiten.Termination
	}

	switch g.scr {
	case scrName:
		g.nameBox.update()
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			g.submitName()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.change(scrLang)
		}

	case scrSettings:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && g.recAct == "" {
			g.change(scrMenu)
			return nil
		}
		if g.recAct != "" {
			if key, ok := grabKey(); ok {
				g.rebindKey(g.recAct, g.recSlot, key)
				g.recAct = ""
				g.saveProfile()
				g.flashMsg = "saved"
				g.flash = 110
			}
		}
		g.trackChips()

	case scrRecords, scrHelp:
		if g.keys.justDown(actBack) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.change(scrMenu)
		}

	case scrMenu:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			g.saveProfile()
			g.quit = true
		}

	case scrGame:
		if g.arc != nil {
			g.arc.update(g)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(rgb(8, 12, 22, 255))
	g.drawAmbient(screen)

	switch g.scr {
	case scrLang:
		g.drawLang(screen)
	case scrName:
		g.drawName(screen)
	case scrMenu:
		g.drawMenu(screen)
	case scrSettings:
		g.drawSettings(screen)
	case scrRecords:
		g.drawRecords(screen)
	case scrHelp:
		g.drawHelp(screen)
	case scrGame:
		if g.arc != nil {
			g.arc.draw(screen, g)
		}
		g.drawTouchControls(screen)
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenW, screenH
}

// RunDesktop запускает игру в отдельном окне (Windows, Linux, macOS).
func RunDesktop() error {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Vizor Space")
	ebiten.SetWindowIcon([]image.Image{makeIcon()})
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return ebiten.RunGame(NewGame())
}
