package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSubmitNamePersistsProfile(t *testing.T) {
	dir := t.TempDir()
	savePath := filepath.Join(dir, "save.dat")
	cfgPath := filepath.Join(dir, "config.cfg")

	previous := configFilePath
	configFilePath = func() string { return cfgPath }
	t.Cleanup(func() { configFilePath = previous })

	g := &Game{
		savePath:  savePath,
		keys:      defaultKeyMap(),
		diff:      DiffHard,
		langID:    LangUK,
		afterName: scrSettings,
	}
	g.nameBox.buf = []byte("  Пилот  ")
	g.submitName()

	if g.name != "Пилот" {
		t.Fatalf("имя не очистилось: %q", g.name)
	}
	if g.scr != scrSettings {
		t.Fatalf("после имени должен открыться scrSettings, получено %v", g.scr)
	}

	lang, name, keys, diff, ok := readSave(savePath)
	if !ok {
		t.Fatal("профиль не сохранился")
	}
	if lang != LangUK || name != "Пилот" || diff != DiffHard {
		t.Fatalf("профиль сохранился криво: lang=%q name=%q diff=%q", lang, name, diff)
	}
	if keys == nil {
		t.Fatal("раскладка не сохранилась")
	}
}

func TestReadSaveRejectsBrokenJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.dat")
	if err := os.WriteFile(path, []byte("{это не json"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, _, _, _, ok := readSave(path); ok {
		t.Fatal("битый JSON не должен считаться валидным профилем")
	}
}

func TestParsers(t *testing.T) {
	if lang, ok := parseLang("UK"); !ok || lang != LangUK {
		t.Fatalf("parseLang(UK) = %q, %v", lang, ok)
	}
	if _, ok := parseLang("klingon"); ok {
		t.Fatal("неизвестный язык не должен парситься")
	}
	if got := parseDiff("HARD"); got != DiffHard {
		t.Fatalf("parseDiff(HARD) = %q", got)
	}
	if got := parseDiff("что-то"); got != DiffNormal {
		t.Fatalf("неизвестная сложность должна стать normal, получено %q", got)
	}
}
