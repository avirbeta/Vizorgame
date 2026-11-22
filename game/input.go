package game

import (
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// NameField — графическое текстовое поле под кириллицу и латиницу.
// Хранит байты UTF-8 (так отдаёт ebiten.AppendInputChars), а для
// операций с символами переводит их в []rune.
type NameField struct {
	buf []byte
	max int
}

func (nf *NameField) update() {
	if nf.max == 0 {
		nf.max = 20
	}

	// затираем последний символ по Backspace
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		rs := []rune(string(nf.buf))
		if len(rs) > 0 {
			rs = rs[:len(rs)-1]
			nf.buf = []byte(string(rs))
		}
	}

	// принимаем всё, что набрал игрок (работает и с раскладками)
	chunk := ebiten.AppendInputChars(nil)
	cleared := make([]rune, 0, len(chunk))
	for _, r := range string(chunk) {
		if r == '\r' || r == '\n' || unicode.IsControl(r) {
			continue
		}
		cleared = append(cleared, r)
	}
	nf.buf = append(nf.buf, []byte(string(cleared))...)

	// не даём набить слишком длинное имя
	if max := nf.max; len([]rune(string(nf.buf))) > max {
		rs := []rune(string(nf.buf))
		rs = rs[:max]
		nf.buf = []byte(string(rs))
	}
}

// name — чистовой вывод поля.
func (nf *NameField) name() string {
	return string(nf.buf)
}

func (nf *NameField) empty() bool {
	return len(nf.buf) == 0
}
