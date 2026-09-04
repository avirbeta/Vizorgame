package main

import (
	"bytes"
	"math"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// Всё звуковое хозяйство крутится на одной частоте дискретизации,
// чтобы не мучиться с ресемплингом: 48 кГц, стерео, 16 бит.
const sampleRate = 48000

// sfxChip — маленькая звуковая «перкуссия»: по пулу плееров на эффект,
// чтобы выстрел, наложенный на выстрел, не хрипел, а спокойно играл
// поверх предыдущего. vol — общая громкость из config.cfg (0..1).
type sfxChip struct {
	ctx   *audio.Context
	vol   float64
	shoot []*audio.Player
	boom  []*audio.Player
}

func newSfx(vol float64) *sfxChip {
	if vol < 0 {
		vol = 0
	}
	if vol > 1 {
		vol = 1
	}
	c := audio.NewContext(sampleRate)
	return &sfxChip{
		ctx:   c,
		vol:   vol,
		shoot: playerPool(c, pewSamples(), 7),
		boom:  playerPool(c, boomSamples(), 5),
	}
}

// PlayShoot — коротенький «пиу».
func (s *sfxChip) PlayShoot() {
	letLoose(s.shoot, 0.55*s.vol)
}

// PlayBoom — развалившийся астероид.
func (s *sfxChip) PlayBoom() {
	letLoose(s.boom, 0.8*s.vol)
}

// playerPool делает n плееров из одного PCM-куска.
func playerPool(c *audio.Context, pcm []byte, n int) []*audio.Player {
	ps := make([]*audio.Player, 0, n)
	for i := 0; i < n; i++ {
		ps = append(ps, audio.NewPlayerFromBytes(c, pcm))
	}
	return ps
}

// letLoose выбирает свободный плеер, выставляет громкость и запускает
// звук с начала.
func letLoose(pool []*audio.Player, vol float64) {
	for _, p := range pool {
		if !p.IsPlaying() {
			p.SetVolume(vol)
			_ = p.Rewind()
			p.Play()
			return
		}
	}
	// все заняты — перебиваем самый первый
	pool[0].SetVolume(vol)
	_ = pool[0].Rewind()
	pool[0].Play()
}

// mkPCM превращает моно-функцию звука в стерео-байты для аудио-контекста.
// gen получает время в секундах и возвращает амплитуду в [-1, 1].
func mkPCM(durSec float64, gen func(t float64) float64) []byte {
	frames := int(durSec * sampleRate)
	var buf bytes.Buffer
	for i := 0; i < frames; i++ {
		t := float64(i) / sampleRate
		s := gen(t)
		if s < -1 {
			s = -1
		}
		if s > 1 {
			s = 1
		}
		v := int16(s * 32767)
		lo, hi := byte(v), byte(v>>8)
		// два канала, у нас всё моно, дублируем
		buf.Write([]byte{lo, hi, lo, hi})
	}
	return buf.Bytes()
}

// pewSamples — «пиу-пиу»: синус с падающей частотой и затуханием.
func pewSamples() []byte {
	const f0, f1 = 760.0, 210.0 // от писклявого к басовитому
	const dur = 0.09

	return mkPCM(dur, func(t float64) float64 {
		k := (f1 - f0) / dur
		ph := 2 * math.Pi * (f0*t + 0.5*k*t*t)
		s := math.Sin(ph)
		// небольшой хрип снизу, чтобы сверхзвук не звучал стерильно
		noise := (rand.Float64()*2 - 1) * 0.08 * (t / dur)
		return (s*0.85 + noise) * math.Exp(-t*16)
	})
}

// boomSamples — «БАХ»: белый шум с резким спадом плюс низкий гул.
func boomSamples() []byte {
	const dur = 0.38

	return mkPCM(dur, func(t float64) float64 {
		env := math.Exp(-t * 9)
		noise := rand.Float64()*2 - 1
		// тяжёлая «нога» внизу
		thump := math.Sin(2 * math.Pi * (85 - 40*t) * t)
		return (noise*0.9 + thump*0.45) * env
	})
}
