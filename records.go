package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const recordCap = 5

// record — одна строка таблицы рекордов.
type record struct {
	Name  string
	Score int
}

// recordsPath — где живёт файл рекордов: рядом с профилем, чтобы носить
// их вместе (и при портативном режиме, и в системной папке).
func recordsPath(save string) string {
	return filepath.Join(filepath.Dir(save), "highscores.dat")
}

// loadRecords читает таблицу. Если своего файла ещё нет, подселяет
// рекорды из папки рядом с игрой — это готовый «сид» первой поставки.
func loadRecords(save string) []record {
	if b, err := os.ReadFile(recordsPath(save)); err == nil {
		if rs := parseRecords(string(b)); len(rs) > 0 {
			return rs
		}
	}
	if b, err := os.ReadFile(filepath.Join(exeDir(), "highscores.dat")); err == nil {
		return parseRecords(string(b))
	}
	return nil
}

// parseRecords разбирает строки вида «Имя|Очки», сортирует и усекает
// до recordCap позиций.
func parseRecords(raw string) []record {
	var rs []record
	for _, ln := range strings.Split(strings.TrimSpace(raw), "\n") {
		s := strings.TrimSpace(ln)
		if s == "" {
			continue
		}
		if i := strings.LastIndexByte(s, '|'); i > 0 {
			name := strings.TrimSpace(s[:i])
			score, err := strconv.Atoi(strings.TrimSpace(s[i+1:]))
			if err == nil && name != "" {
				rs = append(rs, record{Name: name, Score: score})
			}
		}
	}
	orderRecords(rs)
	if len(rs) > recordCap {
		rs = rs[:recordCap]
	}
	return rs
}

// addRecord вносит новый результат в таблицу и пишет файл. Возвращает
// свежую таблицу, чтобы экран проигрыша показал актуальный рекорд.
func addRecord(save, name string, score int) []record {
	rs := loadRecords(save)
	rs = append(rs, record{Name: name, Score: score})
	orderRecords(rs)
	if len(rs) > recordCap {
		rs = rs[:recordCap]
	}

	dir := filepath.Dir(recordsPath(save))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return rs
	}
	var b strings.Builder
	for _, r := range rs {
		fmt.Fprintf(&b, "%s|%d\n", r.Name, r.Score)
	}
	_ = os.WriteFile(recordsPath(save), []byte(b.String()), 0o644)
	return rs
}

func orderRecords(rs []record) {
	sort.Slice(rs, func(i, j int) bool {
		if rs[i].Score != rs[j].Score {
			return rs[i].Score > rs[j].Score
		}
		return rs[i].Name < rs[j].Name
	})
}

// bestScore — лучший результат из таблицы, для показа «Рекорд».
func bestScore(rs []record) int {
	if len(rs) == 0 {
		return 0
	}
	return rs[0].Score
}
