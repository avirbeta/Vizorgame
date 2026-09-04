package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSmokeNewCode(t *testing.T) {
	t.Run("decodeShippedSave", func(t *testing.T) {
		lang, name, km, diff, ok := readSave("save.dat")
		if !ok {
			t.Fatalf("readSave не прочитал shipped save.dat")
		}
		if lang != LangRU || name != "1" {
			t.Fatalf("lang=%q name=%q", lang, name)
		}
		if diff != DiffNormal {
			t.Fatalf("difficulty=%q", diff)
		}
		if km == nil {
			t.Fatal("кладка nil")
		}
	})

	t.Run("parseRecords", func(t *testing.T) {
		rs := parseRecords("Pilot|300\nВитя|150\n")
		if len(rs) != 2 || rs[0].Name != "Pilot" || rs[0].Score != 300 {
			t.Fatalf("рекорды разобрались криво: %+v", rs)
		}
	})

	t.Run("addRecordRoundTrip", func(t *testing.T) {
		dir := t.TempDir()
		save := filepath.Join(dir, "save.dat")
		rs := addRecord(save, "Тест", 42)
		if bestScore(rs) != 42 {
			t.Fatalf("лучший не 42: %+v", rs)
		}
		rs2 := loadRecords(save)
		if len(rs2) != 1 || rs2[0].Score != 42 {
			t.Fatalf("перечитанный файл крив: %+v", rs2)
		}
	})

	t.Run("configFile", func(t *testing.T) {
		dir := t.TempDir()
		cfgPath := filepath.Join(dir, "config.cfg")
		if err := os.WriteFile(cfgPath, []byte("lang = uk\nvolume = 50\nportable = 1\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		defer func(prev func() string) { configFilePath = prev }(configFilePath)
		configFilePath = func() string { return cfgPath }
		cfg := loadConfig()
		if !cfg.hasLang || cfg.lang != LangUK || cfg.volume != 0.5 || !cfg.portable {
			t.Fatalf("config разобран криво: %+v", cfg)
		}
	})

	t.Run("keysJsonRoundTrip", func(t *testing.T) {
		b, _ := json.Marshal(defaultKeyMap())
		km := decodeKeys(string(b))
		if len(km[actShoot]) == 0 {
			t.Fatal("раскладка из JSON не восстановилась")
		}
	})
}
