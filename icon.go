package main

import (
	"image"
	"image/color"
	"math"
)

// makeIcon рисует эмблемку игры прямо в памяти, ни одного .png
// в папке не появляется. Дальше она уходит в SetWindowIcon.
func makeIcon() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 128, 128))

	// тёмный диск-подложка, чтобы иконка читалась на любом фоне
	fillCircle(img, 64, 64, 56, color.RGBA{20, 26, 40, 255})
	// стальная окантовка диска
	strokeCircle(img, 64, 64, 54, 3, color.RGBA{92, 134, 164, 255})

	// звёздочки вокруг корабля
	starDot(img, 38, 42, 2.2, color.RGBA{232, 240, 250, 235})
	starDot(img, 94, 34, 1.6, color.RGBA{232, 240, 250, 200})
	starDot(img, 96, 84, 2.0, color.RGBA{232, 240, 250, 215})
	starDot(img, 32, 86, 1.4, color.RGBA{232, 240, 250, 185})

	// корпус — широкая «стрела»
	fillTriangleA(img, 64, 30, 38, 88, 90, 88, color.RGBA{150, 220, 255, 255})
	// нос посветлее, чтобы появился объём
	fillTriangleA(img, 64, 30, 55, 52, 73, 52, color.RGBA{226, 246, 255, 255})
	// кабина
	fillCircle(img, 64, 62, 7, color.RGBA{250, 252, 255, 255})
	// пламя из хвоста
	fillTriangleA(img, 64, 88, 52, 106, 76, 106, color.RGBA{255, 178, 77, 255})
	fillTriangleA(img, 64, 90, 57, 101, 71, 101, color.RGBA{255, 214, 130, 255})

	// пара астероидов по бокам — намёк на интерактив
	fillCircle(img, 24, 58, 9, color.RGBA{150, 118, 98, 255})
	fillCircle(img, 21, 55, 3, color.RGBA{110, 84, 70, 255})
	fillCircle(img, 104, 66, 8, color.RGBA{146, 152, 122, 255})
	fillCircle(img, 101, 63, 3, color.RGBA{104, 110, 88, 255})

	return img
}

// fillCircle — грубоватая, но быстрая заливка круга.
func fillCircle(img *image.RGBA, cx, cy, r float64, c color.RGBA) {
	for y := int(cy-r) - 1; y <= int(cy+r)+1; y++ {
		if y < 0 || y >= img.Bounds().Dy() {
			continue
		}
		for x := int(cx-r) - 1; x <= int(cx+r)+1; x++ {
			if x < 0 || x >= img.Bounds().Dx() {
				continue
			}
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r*r {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

// strokeCircle — кольцо заданной толщины.
func strokeCircle(img *image.RGBA, cx, cy, r, th float64, c color.RGBA) {
	for y := int(cy - r - th); y <= int(cy+r+th)+1; y++ {
		if y < 0 || y >= img.Bounds().Dy() {
			continue
		}
		for x := int(cx - r - th); x <= int(cx+r+th)+1; x++ {
			if x < 0 || x >= img.Bounds().Dx() {
				continue
			}
			d := math.Hypot(float64(x)+0.5-cx, float64(y)+0.5-cy)
			if d >= r-th && d <= r {
				img.SetRGBA(x, y, c)
			}
		}
	}
}

// fillTriangleA заливает треугольник по барицентрическим координатам.
func fillTriangleA(img *image.RGBA, ax, ay, bx, by, cx, cy float64, col color.RGBA) {
	minX := int(math.Min(math.Min(ax, bx), cx)) - 1
	maxX := int(math.Max(math.Max(ax, bx), cx)) + 1
	minY := int(math.Min(math.Min(ay, by), cy)) - 1
	maxY := int(math.Max(math.Max(ay, by), cy)) + 1
	for y := minY; y <= maxY; y++ {
		if y < 0 || y >= img.Bounds().Dy() {
			continue
		}
		for x := minX; x <= maxX; x++ {
			if x < 0 || x >= img.Bounds().Dx() {
				continue
			}
			px, py := float64(x)+0.5, float64(y)+0.5
			if pointInTri(px, py, ax, ay, bx, by, cx, cy) {
				img.SetRGBA(x, y, col)
			}
		}
	}
}

func pointInTri(px, py, ax, ay, bx, by, cx, cy float64) bool {
	d := (by-cy)*(ax-cx) + (cx-bx)*(ay-cy)
	wa := ((by-cy)*(px-cx) + (cx-bx)*(py-cy)) / d
	wb := ((cy-ay)*(px-cx) + (ax-cx)*(py-cy)) / d
	wc := 1 - wa - wb
	return wa >= 0 && wb >= 0 && wc >= 0
}

// starDot — звёздочка: серединка и четыре лучика крестиком.
func starDot(img *image.RGBA, x, y, r float64, c color.RGBA) {
	fillCircle(img, x, y, r, c)
	n := int(r*3) + 2
	dirs := [][2]float64{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
	for i := 1; i <= n; i++ {
		for _, s := range dirs {
			xx, yy := int(x+s[0]*float64(i)), int(y+s[1]*float64(i))
			if xx >= 0 && yy >= 0 && xx < 128 && yy < 128 {
				img.SetRGBA(xx, yy, c)
			}
		}
	}
}
