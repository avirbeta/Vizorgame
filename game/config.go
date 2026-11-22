package game

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// config — содержимое config.cfg: язык мастера, громкость и портативный режим.
type config struct {
	hasLang  bool
	lang     LangID
	volume   float64
	portable bool
}

func defaultConfig() config {
	return config{volume: 0.7}
}

// configFilePath вынесен в переменную, чтобы тесты могли подменить путь.
var configFilePath = func() string {
	return filepath.Join(exeDir(), "config.cfg")
}

func loadConfig() config {
	cfg := defaultConfig()

	file, err := os.Open(configFilePath())
	if err != nil {
		return cfg
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.ToLower(strings.TrimSpace(key))
		value = strings.TrimSpace(value)

		switch key {
		case "lang":
			if lang, ok := parseLang(value); ok {
				cfg.lang = lang
				cfg.hasLang = true
			}
		case "volume":
			if n, err := strconv.Atoi(value); err == nil {
				if n < 0 {
					n = 0
				}
				if n > 100 {
					n = 100
				}
				cfg.volume = float64(n) / 100
			}
		case "portable":
			cfg.portable = value == "1" ||
				strings.EqualFold(value, "true") ||
				strings.EqualFold(value, "yes")
		}
	}
	return cfg
}

func saveConfig(cfg config) bool {
	lang := cfg.lang
	if lang == "" {
		lang = LangEN
	}
	volume := int(cfg.volume*100 + 0.5)
	portable := 0
	if cfg.portable {
		portable = 1
	}
	body := strings.Join([]string{
		"# Vizor Space — настройки",
		"lang = " + string(lang),
		"volume = " + strconv.Itoa(volume),
		"portable = " + strconv.Itoa(portable),
		"",
	}, "\n")

	path := configFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return false
	}
	return os.WriteFile(path, []byte(body), 0o644) == nil
}

// exeDir — папка, рядом с которой лежит бинарник (для portable-режима).
func exeDir() string {
	path, err := os.Executable()
	if err != nil || path == "" {
		return "."
	}
	if resolved, err := filepath.EvalSymlinks(path); err == nil {
		path = resolved
	}
	return filepath.Dir(path)
}

// dataDir — системная папка профиля, когда portable выключен.
func dataDir() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "MyCustomGame")
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, ".mycustomgame")
	}
	return "."
}

func resolveSavePath(cfg config) string {
	if cfg.portable {
		return filepath.Join(exeDir(), "save.dat")
	}
	return filepath.Join(dataDir(), "save.dat")
}
