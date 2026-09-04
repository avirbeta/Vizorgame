package main

import (
	"bytes"
	_ "embed"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// DejaVu Sans отлично покрывает кириллицу, украинские буквы и испанский,
// поэтому вшиваем его прямо в exe — финальному игроку ни одного шрифта
// ставить не придётся.
//
//go:embed assets/font/DejaVuSans.ttf
var fontTTF []byte

var gFontSource *text.GoTextFaceSource

// whiteImg — белый пиксель; с него берём текстуру для заливки треугольников.
var whiteImg *ebiten.Image

func init() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(fontTTF))
	if err != nil {
		panic(err)
	}
	gFontSource = src

	whiteImg = ebiten.NewImage(1, 1)
	whiteImg.Fill(color.White)
}

// face — короткий конструктор шрифтового лица нужного размера.
func face(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: gFontSource, Size: size}
}
