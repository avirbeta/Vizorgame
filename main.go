package main

import (
	"image"
	"image/color"
	"log"
	"math/rand"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenW = 1280
	screenH = 720
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
}

func newGame() *Game {
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

	for _, list := range [][]*Button{
		g.langBtns, g.menuBtns, g.nameBtn, g.setBtns,
		g.diffBtns, g.resetBtn, g.backBtns,
	} {
		for _, button := range list {
			button.update()
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
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenW, screenH
}

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Vizor Space")
	ebiten.SetWindowIcon([]image.Image{makeIcon()})
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
