package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gitlab.com/slon/shad-go/gitfame/internal/formatters"
	"gitlab.com/slon/shad-go/gitfame/internal/models"
)

type BlameLine = models.BlameLine

func CollectFiles(repositoryPath string, languages []string, extensions []string, exclude []string, restrictTo []string, revision string, languageMap *map[string][]string) ([]string, error) {
	extSet := make(map[string]struct{})
	for _, ext := range extensions {
		norm := normalizeExt(ext)
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
			norm := normalizeExt(ext)
			if norm != "" {
				extSet[norm] = struct{}{}
			}
		}
	}

	out, err := exec.Command("git", "-C", repositoryPath, "ls-tree", "-r", revision).Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(out))
	var files []string
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}

		header := strings.Fields(parts[0])
		if len(header) < 2 || header[1] != "blob" {
			continue
		}

		filePath := strings.TrimSpace(parts[1])
		if filePath == "" {
			continue
		}

		if len(extSet) > 0 {
			ext := strings.ToLower(filepath.Ext(filePath))
			if _, ok := extSet[ext]; !ok {
				continue
			}
		}

		if len(restrictTo) > 0 {
			matched := false
			for _, pattern := range restrictTo {
				pattern = strings.TrimSpace(pattern)
				if pattern == "" {
					continue
				}
				ok, _ := filepath.Match(pattern, filePath)
				if ok {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		excluded := false
		for _, pattern := range exclude {
			pattern = strings.TrimSpace(pattern)
			if pattern == "" {
				continue
			}
			ok, _ := filepath.Match(pattern, filePath)
			if ok {
				excluded = true
				break
			}
		}
		if excluded {
			continue
		}

		files = append(files, filePath)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

// dfs по файлам
func CollectBlame(repositoryPath string, revision string, files []string, commiter bool) ([]BlameLine, error) {
	res := make([]BlameLine, 0)
	for _, filePath := range files {
		file, err := Blame(repositoryPath, revision, filePath, commiter)
		if err != nil {
			return nil, err
		}
		res = append(res, file...)
	}
	return res, nil
}

// для 1 файла
func Blame(repositoryPath string, revision string, filePath string, commiter bool) ([]BlameLine, error) {
	out, err := exec.Command("git", "-C", repositoryPath, "blame", "--porcelain", revision, "--", filePath).Output()
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(out))
	type authorInfo struct {
		commits map[string]struct{}
		lines   int
	}
	info := make(map[string]*authorInfo)

	var currentCommit string
	var currentAuthor string
	var prevline string

	for scanner.Scan() {
		line := scanner.Text()
		if currentCommit == "" && currentAuthor == "" {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				currentCommit = parts[0]
			}
			prevline = line
			continue
		}

		if strings.HasPrefix(line, "author ") {
			currentAuthor = strings.TrimSpace(strings.TrimPrefix(line, "author "))
			parts := strings.Fields(prevline)
			if len(parts) > 0 {
				currentCommit = parts[0]
			}
			if !commiter {
				if _, exists := info[currentAuthor]; !exists {
					info[currentAuthor] = &authorInfo{commits: make(map[string]struct{})}
				}
				info[currentAuthor].commits[currentCommit] = struct{}{}
			}
			prevline = line
			continue
		}

		if commiter && strings.HasPrefix(line, "committer ") {
			currentAuthor = strings.TrimSpace(strings.TrimPrefix(line, "committer "))
			parts := strings.Fields(prevline)
			if len(parts) > 0 {
				currentCommit = parts[0]
			}
			if _, exists := info[currentAuthor]; !exists {
				info[currentAuthor] = &authorInfo{commits: make(map[string]struct{})}
			}
			info[currentAuthor].commits[currentCommit] = struct{}{}
			prevline = line
			continue
		}

		if strings.HasPrefix(line, "\t") {
			if currentAuthor == "" {
				prevline = line
				continue
			}
			if _, exists := info[currentAuthor]; !exists {
				info[currentAuthor] = &authorInfo{commits: make(map[string]struct{})}
			}
			info[currentAuthor].lines++
			prevline = line
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	var result []BlameLine
	for author, a := range info {
		result = append(result, BlameLine{
			Name:    author,
			Lines:   a.lines,
			Commits: len(a.commits),
			Files:   1,
		})
	}

	return result, nil
}

func normalizeExt(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	if !strings.HasPrefix(s, ".") {
		s = "." + s
	}
	return strings.ToLower(s)
}
