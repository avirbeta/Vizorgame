package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// Difficulty — выбранный режим сложности.
type Difficulty string

const (
	DiffEasy   Difficulty = "easy"
	DiffNormal Difficulty = "normal"
	DiffHard   Difficulty = "hard"
)

func parseDiff(value string) Difficulty {
	switch Difficulty(strings.ToLower(strings.TrimSpace(value))) {
	case DiffEasy:
		return DiffEasy
	case DiffHard:
		return DiffHard
	default:
		return DiffNormal
	}
}

func parseLang(value string) (LangID, bool) {
	switch LangID(strings.ToLower(strings.TrimSpace(value))) {
	case LangEN, LangRU, LangUK, LangES:
		return LangID(strings.ToLower(strings.TrimSpace(value))), true
	}
	return "", false
}

// saveData — формат save.dat. Клавиши храним отдельным JSON-блоком,
// чтобы совместимость с encodeKeys/decodeKeys осталась прежней.
type saveData struct {
	Lang       string          `json:"lang"`
	Name       string          `json:"name"`
	Keys       json.RawMessage `json:"keys,omitempty"`
	Difficulty string          `json:"difficulty"`
}

// readSave читает профиль пилота. ok=false только если файла нет
// или он полностью нечитаем — частично битый профиль чинится значениями
// по умолчанию.
func readSave(path string) (LangID, string, KeyMap, Difficulty, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return LangEN, "", defaultKeyMap(), DiffNormal, false
	}

	var data saveData
	if err := json.Unmarshal(raw, &data); err != nil {
		return LangEN, "", defaultKeyMap(), DiffNormal, false
	}

	lang, ok := parseLang(data.Lang)
	if !ok {
		lang = LangEN
	}
	keys := decodeKeys(string(data.Keys))
	name := strings.TrimSpace(data.Name)
	return lang, name, keys, parseDiff(data.Difficulty), true
}

// writeSave атомарно сохраняет профиль: сначала пишем .tmp, затем
// переименовываем, чтобы внезапное выключение не оставило обрезанный JSON.
func writeSave(path string, lang LangID, name string, keys KeyMap, diff Difficulty) bool {
	encodedKeys := encodeKeys(keys)
	if encodedKeys == "" {
		encodedKeys = encodeKeys(defaultKeyMap())
	}
	data := saveData{
		Lang:       string(lang),
		Name:       name,
		Keys:       json.RawMessage(encodedKeys),
		Difficulty: string(diff),
	}
	body, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return false
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return false
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return false
	}
	return true
}
