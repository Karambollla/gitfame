package formatters

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func LoadLanguageMap(repositoryPath string) (*map[string][]string, error) {
	cfgPath, err := resolveLanguageConfigPath(repositoryPath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var objectMap map[string][]string
	if err := json.Unmarshal(data, &objectMap); err == nil {
		n := normalizeLanguageMap(objectMap)
		return &n, nil
	}

	type languageConfigEntry struct {
		Name       string   `json:"name"`
		Extensions []string `json:"extensions"`
	}

	var entries []languageConfigEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	raw := make(map[string][]string, len(entries))
	for _, entry := range entries {
		if strings.TrimSpace(entry.Name) == "" || len(entry.Extensions) == 0 {
			continue
		}
		raw[entry.Name] = append(raw[entry.Name], entry.Extensions...)
	}

	n := normalizeLanguageMap(raw)
	return &n, nil
}

func resolveLanguageConfigPath(repositoryPath string) (string, error) {
	if strings.TrimSpace(repositoryPath) != "" {
		candidate := filepath.Join(repositoryPath, "configs", "language_extensions.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if candidate, ok := findConfigPathUpward(wd); ok {
			return candidate, nil
		}
	}

	if exe, err := os.Executable(); err == nil {
		if candidate, ok := findConfigPathUpward(filepath.Dir(exe)); ok {
			return candidate, nil
		}
	}

	return "", errors.New("language configuration not found")
}

func findConfigPathUpward(startDir string) (string, bool) {
	dir := filepath.Clean(startDir)
	for {
		candidate := filepath.Join(dir, "configs", "language_extensions.json")
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, true
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func normalizeLanguageMap(langMap map[string][]string) map[string][]string {
	norm := make(map[string][]string, len(langMap))
	for lang, exts := range langMap {
		key := strings.ToLower(strings.TrimSpace(lang))
		if key == "" {
			continue
		}

		out := make([]string, 0, len(exts))
		for _, e := range exts {
			e = strings.TrimSpace(e)
			if e == "" {
				continue
			}
			if !strings.HasPrefix(e, ".") {
				e = "." + e
			}
			out = append(out, strings.ToLower(e))
		}
		norm[key] = out
	}

	return norm
}

func ExtensionsForLanguage(language string, languageMap *map[string][]string) []string {
	if languageMap == nil {
		return nil
	}

	key := strings.ToLower(strings.TrimSpace(language))
	if key == "" {
		return nil
	}

	exts, ok := (*languageMap)[key]
	if !ok {
		return nil
	}

	out := make([]string, 0, len(exts))
	out = append(out, exts...)
	return out
}
