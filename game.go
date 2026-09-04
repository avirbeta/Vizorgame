package main

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// slug — заряд нашего корабля.
type slug struct{ x, y float64 }

// blob — астероид, который летит сверху на бедного пилота.
type blob struct {
	x, y, r     float64
	vx, vy      float64
	spin, angle float64
}

// spark — фоновая звёздочка (или камешек для меню).
type spark struct {
	x, y, r float64
	a       uint8 // прозрачность
}

// Arcade — вся внутренность космо-аркады: корабль, астероиды, счёт.
type Arcade struct {
	px, py float64 // позиция корабля
	tilt   float64 // крен при движении, для красоты

	bullets  []*slug
	rocks    []*blob
	sparkles []*spark

	score, lives int
	kills        int // сбитые глыбы, из них растёт уровень
	level        int // текущий уровень (бесконечная лесенка)
	lvlMsg       int // таймер надписи «Уровень N»
	cooldown     int // пауза между выстрелами
	spawn        int // таймер появления астероида
	invuln       int // секунды бессмертия после ушиба
	shake        int // тряска экрана при взрыве

	over   bool
	paused bool // пауза: Esc или P
}

func newArcade() *Arcade {
	a := &Arcade{
		px:     screenW / 2,
		py:     screenH - 90,
		lives:  3,
		level:  1,
		spawn:  60,
		invuln: 0,
	}
	// сразу раскладываем немножко звёзд
	for i := 0; i < 70; i++ {
		a.sparkles = append(a.sparkles, &spark{
			x: rand.Float64() * screenW,
			y: rand.Float64() * screenH,
			r: 1 + rand.Float64()*2,
			a: uint8(60 + rand.Intn(120)),
		})
	}
	return a
}

// leftRight читает зажатые клавиши движения через раскладку игрока.
func leftRight(km KeyMap) float64 {
	dx := 0.0
	if km.held(actLeft) {
		dx--
	}
	if km.held(actRight) {
		dx++
	}
	return dx
}

func (a *Arcade) update(g *Game) {
	if a.over {
		// финальный экран ждёт действий игрока
		if g.keys.justDown(actShoot) {
			a.reset()
		}
		if g.keys.justDown(actBack) {
			g.change(scrMenu)
		}
		return
	}

	// пауза на Esc или P: замираем, но слушаем клавиши
	if inpututil.IsKeyJustPressed(ebiten.KeyP) || g.keys.justDown(actBack) {
		a.paused = !a.paused
	}
	if a.paused {
		return
	}

	// звёзды ползут вниз — имитация движения
	for _, s := range a.sparkles {
		s.y += 0.6 * (1 + s.r)
		if s.y > screenH {
			s.y = -4
			s.x = rand.Float64() * screenW
		}
	}

	// движение корабля
	dir := leftRight(g.keys)
	speed := 6.5
	a.px += dir * speed
	if a.px < 24 {
		a.px = 24
	}
	if a.px > screenW-24 {
		a.px = screenW - 24
	}
	a.tilt += (dir - a.tilt) * 0.2

	// стрельба
	a.cooldown--
	if g.keys.held(actShoot) && a.cooldown <= 0 {
		a.bullets = append(a.bullets,
			&slug{x: a.px, y: a.py - 24},
			&slug{x: a.px, y: a.py - 24},
		)
		// маленькое разведение залпа
		a.bullets[len(a.bullets)-2].x -= 9
		a.bullets[len(a.bullets)-1].x += 9
		a.cooldown = 12
		g.sfx.PlayShoot()
	}

	// спавн астероидов. Темп зависит от сложности и от того, на каком
	// уровне пилот удержался.
	diff := g.diff
	spawnGap := 52 + rand.Intn(30)
	spdMul := 1.0
	switch diff {
	case DiffEasy:
		spawnGap += 28
		spdMul = 0.78
	case DiffHard:
		spawnGap -= 16
		spdMul = 1.3
	}
	spawnGap -= (a.level - 1) * 3
	if spawnGap < 18 {
		spawnGap = 18
	}

	a.spawn--
	if a.spawn <= 0 {
		r := 16 + rand.Float64()*18
		a.rocks = append(a.rocks, &blob{
			x:     rand.Float64()*(screenW-2*r) + r,
			y:     -30,
			r:     r,
			vy:    (1.6 + float64(a.score)/260 + rand.Float64()) * spdMul,
			vx:    (-0.8 + rand.Float64()*1.6) * spdMul,
			angle: rand.Float64() * 6.28,
			spin:  -0.05 + rand.Float64()*0.1,
		})
		a.spawn = spawnGap
	}

	// снаряды летят вверх
	alive := a.bullets[:0]
	for _, b := range a.bullets {
		b.y -= 14
		if b.y > -20 {
			alive = append(alive, b)
		}
	}
	a.bullets = alive

	a.invuln--

	// астероиды летят вниз и вертятся
	bRocks := a.rocks[:0]
	for _, r := range a.rocks {
		r.x += r.vx
		r.y += r.vy
		r.angle += r.spin
		if r.x < -r.r {
			r.x = screenW + r.r
		}
		if r.x > screenW+r.r {
			r.x = -r.r
		}
		if r.y > screenH+40 {
			continue // улетел за борт, не больно
		}
		bRocks = append(bRocks, r)
	}
	a.rocks = bRocks

	// попали ли снаряды
	hit := func(s *slug, r *blob) bool {
		dx := s.x - r.x
		dy := s.y - r.y
		return dx*dx+dy*dy <= r.r*r.r
	}
	keptBullets := a.bullets[:0]
	keptRocks := a.rocks[:0]
	for _, r := range a.rocks {
		took := false
		for _, s := range a.bullets {
			if hit(s, r) {
				a.score += 10
				a.kills++
				a.shake = 8
				g.sfx.PlayBoom()
				took = true
				break
			}
		}
		if !took {
			keptRocks = append(keptRocks, r)
		}
	}
	for _, s := range a.bullets {
		dead := false
		for _, r := range a.rocks {
			if hit(s, r) {
				dead = true
				break
			}
		}
		if !dead {
			keptBullets = append(keptBullets, s)
		}
	}
	a.bullets = keptBullets
	a.rocks = keptRocks

	// столкновение корабля с астероидом
	if a.invuln <= 0 {
		for _, r := range a.rocks {
			dx := a.px - r.x
			dy := a.py - r.y
			if dx*dx+dy*dy <= (r.r+18)*(r.r+18) {
				a.lives--
				a.invuln = 120
				a.shake = 14
				g.sfx.PlayBoom()
				r.y = screenH + 99 // «выпихиваем» глыбу за экран
				break
			}
		}
	}

	if a.shake > 0 {
		a.shake--
	}

	// уровень растёт через каждые 10 сбитых глыб
	nl := 1 + a.kills/10
	if nl != a.level {
		a.level = nl
		a.lvlMsg = 120
		g.sfx.PlayBoom()
		// каждые четыре уровня — бонусная жизнь, но не больше шести
		if a.level%4 == 0 && a.lives < 6 {
			a.lives++
		}
	}
	if a.lvlMsg > 0 {
		a.lvlMsg--
	}

	// итог: только поражение. Уровни бесконечные.
	if a.lives <= 0 {
		a.over = true
	}
	if a.over {
		// партия закончилась — вносим результат в таблицу рекордов
		g.best = bestScore(addRecord(g.savePath, g.name, a.score))
		g.recs = loadRecords(g.savePath)
	}
}

// reset возвращает аркаду в стартовое состояние для новой партии.
func (a *Arcade) reset() {
	a.score = 0
	a.lives = 3
	a.kills = 0
	a.level = 1
	a.lvlMsg = 0
	a.px = screenW / 2
	a.py = screenH - 90
	a.bullets = nil
	a.rocks = nil
	a.cooldown = 0
	a.spawn = 60
	a.invuln = 0
	a.shake = 0
	a.paused = false
	a.over = false
	for _, s := range a.sparkles {
		s.y = rand.Float64() * screenH
	}
}

func (a *Arcade) draw(screen *ebiten.Image, g *Game) {
	// фоновые звёзды
	for _, s := range a.sparkles {
		vector.DrawFilledCircle(screen, float32(s.x), float32(s.y), float32(s.r), color.RGBA{150, 180, 220, s.a}, true)
	}

	// тряска экрана
	ox, oy := 0.0, 0.0
	if a.shake > 0 {
		ox = rand.Float64()*6 - 3
		oy = rand.Float64()*6 - 3
	}

	// астероиды — кособокие глыбы с кратерами
	for _, r := range a.rocks {
		drawRock(screen, r, ox, oy)
	}

	// снаряды — короткие светящиеся тире
	for _, b := range a.bullets {
		vector.DrawFilledRect(screen, float32(b.x-2+ox), float32(b.y-9+oy), 4, 14, clrBullet, true)
		clrBulletBurst := rgb(255, 235, 180, 120)
		vector.DrawFilledCircle(screen, float32(b.x+ox), float32(b.y+oy), 5, clrBulletBurst, true)
	}

	// корабль
	var shipColor color.Color = clrShip
	blinky := (g.nowTicks/8)%2 == 0
	if a.invuln > 0 && blinky {
		shipColor = rgb(140, 220, 255, 90)
	}
	sx, sy := float32(a.px+ox), float32(a.py+oy)
	// крылья + нос — одна примитивная «галка»
	fillTriangle(screen, sx, sy-26, sx-20, sy+18, sx, sy+6, shipColor)
	fillTriangle(screen, sx, sy-26, sx+20, sy+18, sx, sy+6, shipColor)
	// кабина
	vector.DrawFilledCircle(screen, sx, sy-4, 5, rgb(240, 248, 255, 255), true)
	// маленький хвостовой огонёк
	vector.DrawFilledCircle(screen, sx, sy+20, 4, rgb(255, 170, 60, 220), true)

	// HUD
	drawString(screen, fmt.Sprintf("%s: %d", g.t("score"), a.score), 120, 30, 26, clrText)
	drawString(screen, fmt.Sprintf("%s: %d", g.t("level"), a.level), screenW/2, 30, 26, clrDim)
	for i := 0; i < a.lives; i++ {
		li, lj := float64(screenW-200+i*36), 22.0
		fillTriangle(screen, float32(li), float32(lj+6), float32(li-8), float32(lj+16), float32(li+8), float32(lj+16), clrShip)
	}
	drawString(screen, g.t("hintMove"), screenW/2, screenH-24, 15, clrDim)

	// пауза: приглушаем поле и пишем подсказку
	if a.paused {
		screen.Fill(rgb(4, 6, 12, 150))
		drawStringShadow(screen, g.t("pause"), screenW/2, 300, 54, clrAccent)
		drawString(screen, fmt.Sprintf("%s: %d", g.t("score"), a.score), screenW/2, 366, 26, clrText)
		drawString(screen, g.t("pauseHint"), screenW/2, 420, 24, clrDim)
	}

	// всплывающая надпись о новом уровне
	if a.lvlMsg > 0 {
		a2 := 220 * float64(a.lvlMsg) / 120
		if a2 > 255 {
			a2 = 255
		}
		drawStringShadow(screen, fmt.Sprintf("%s %d", g.t("level"), a.level), screenW/2, 220, 46, rgb(255, 214, 130, uint8(a2)))
	}

	// стекло проигрыша
	if a.over {
		screen.Fill(rgb(4, 6, 12, 160))
		drawStringShadow(screen, fmt.Sprintf(g.t("lose"), g.name), screenW/2, 250, 52, clrDanger)
		drawString(screen, fmt.Sprintf("%s: %d", g.t("score"), a.score), screenW/2, 330, 28, clrText)
		if g.best > 0 {
			drawString(screen, fmt.Sprintf("%s: %d", g.t("record"), g.best), screenW/2, 370, 22, clrDim)
		}
		drawString(screen, g.t("again"), screenW/2, 400, 22, clrDim)
		drawString(screen, g.t("backMenu"), screenW/2, 440, 22, clrDim)
	}
}

// drawRock рисует астероид с «молниями» и кратером.
func drawRock(screen *ebiten.Image, r *blob, ox, oy float64) {
	cx, cy := float32(r.x+ox), float32(r.y+oy)
	rad := float32(r.r)

	body := rgb(128, 100, 84, 255)
	shade := rgb(90, 68, 56, 255)
	hi := rgb(158, 128, 108, 255)

	// тень под брюхом
	vector.DrawFilledCircle(screen, cx+rad*0.28, cy+rad*0.24, rad, shade, true)
	// основная глыба
	vector.DrawFilledCircle(screen, cx, cy, rad, body, true)
	// блик
	vector.DrawFilledCircle(screen, cx-rad*0.3, cy-rad*0.3, rad*0.55, hi, true)
	// кратер
	vector.DrawFilledCircle(screen, cx+rad*0.25, cy-radiusOffset(rad), rad*0.28, rgb(70, 54, 44, 255), true)
	vector.DrawFilledCircle(screen, cx+rad*0.05, cy+rad*0.4, rad*0.2, rgb(70, 54, 44, 255), true)
}

// radiusOffset — лёгкая случайность положения кратера на глыбе.
func radiusOffset(r float32) float32 {
	return r * (0.5 + 0.3*float32(math.Abs(float64(math.Sin(float64(r*13))))))
}

// fillTriangle — дешёвая заливка треугольника заданным цветом через Path.
func fillTriangle(screen *ebiten.Image, ax, ay, bx, by, cx, cy float32, col color.Color) {
	var p vector.Path
	p.MoveTo(ax, ay)
	p.LineTo(bx, by)
	p.LineTo(cx, cy)
	p.Close()
	vs, is := p.AppendVerticesAndIndicesForFilling(nil, nil)
	cr, cg, cb, ca := col.RGBA()
	for i := range vs {
		vs[i].ColorR = float32(cr) / 65535
		vs[i].ColorG = float32(cg) / 65535
		vs[i].ColorB = float32(cb) / 65535
		vs[i].ColorA = float32(ca) / 65535
	}
	screen.DrawTriangles(vs, is, whiteImg, nil)
}
