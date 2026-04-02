package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gitlab.com/slon/shad-go/gitfame/internal/formatters"
	"gitlab.com/slon/shad-go/gitfame/internal/util"
)

func fileMatchesFilters(filePath string, extSet map[string]struct{}, exclude []string, restrictTo []string) bool {
	if len(extSet) > 0 {
		ext := strings.ToLower(filepath.Ext(filePath))
		if _, ok := extSet[ext]; !ok {
			return false
		}
	}

	if len(restrictTo) > 0 && !matchAny(restrictTo, filePath) {
		return false
	}

	if matchAny(exclude, filePath) {
		return false
	}

	return true
}

func matchAny(patterns []string, filePath string) bool {
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		ok, _ := filepath.Match(p, filePath)
		if ok {
			return true
		}
	}
	return false
}

func buildExtSet(extensions []string, languages []string, languageMap *map[string][]string) map[string]struct{} {
	extSet := make(map[string]struct{})
	for _, ext := range extensions {
		norm := util.NormalizeExt(ext)
		if norm != "" {
			extSet[norm] = struct{}{}
		}
	}
	for _, lang := range languages {
		langExts := formatters.ExtensionsForLanguage(lang, languageMap)
		if len(langExts) == 0 {
			if strings.TrimSpace(lang) != "" {
				fmt.Fprintln(os.Stderr, "warning: unknown language:", lang)
			}
			continue
		}
		for _, ext := range langExts {
			norm := util.NormalizeExt(ext)
			if norm != "" {
				extSet[norm] = struct{}{}
			}
		}
	}
	return extSet
}
