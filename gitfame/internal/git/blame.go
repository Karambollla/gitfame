package git

import (
	"bufio"
	"bytes"
	"os/exec"
	"strconv"
	"strings"

	"gitlab.com/slon/shad-go/gitfame/internal/models"
)

type BlameLine = models.BlameLine

type authorInfo struct {
	commits map[string]struct{}
	lines   int
}
type totalAuthorInfo struct {
	commits, files map[string]struct{}
	lines          int
}

func CollectFiles(repositoryPath string, languages []string, extensions []string, exclude []string, restrictTo []string, revision string, languageMap *map[string][]string) ([]string, error) {
	extSet := buildExtSet(extensions, languages, languageMap)

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

		if !fileMatchesFilters(filePath, extSet, exclude, restrictTo) {
			continue
		}

		files = append(files, filePath)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return files, nil
}

func CollectBlame(repositoryPath string, revision string, files []string, commiter bool) ([]BlameLine, error) {
	total := make(map[string]*totalAuthorInfo)
	for _, filePath := range files {
		fileInfo, err := collectFileBlame(repositoryPath, revision, filePath, commiter)
		if err != nil {
			return nil, err
		}
		for author, info := range fileInfo {
			a := total[author]
			if a == nil {
				a = &totalAuthorInfo{commits: map[string]struct{}{}, files: map[string]struct{}{}}
				total[author] = a
			}
			a.lines += info.lines
			for commit := range info.commits {
				a.commits[commit] = struct{}{}
			}
			a.files[filePath] = struct{}{}
		}
	}
	res := make([]BlameLine, 0, len(total))
	for author, info := range total {
		res = append(res, BlameLine{Name: author, Lines: info.lines, Commits: len(info.commits), Files: len(info.files)})
	}
	return res, nil
}

func Blame(repositoryPath string, revision string, filePath string, commiter bool) ([]BlameLine, error) {
	fileInfo, err := collectFileBlame(repositoryPath, revision, filePath, commiter)
	if err != nil {
		return nil, err
	}
	result := make([]BlameLine, 0, len(fileInfo))
	for author, info := range fileInfo {
		result = append(result, BlameLine{Name: author, Lines: info.lines, Commits: len(info.commits), Files: 1})
	}
	return result, nil
}

func collectFileBlame(repositoryPath string, revision string, filePath string, commiter bool) (map[string]*authorInfo, error) {
	out, err := exec.Command("git", "-C", repositoryPath, "blame", "--porcelain", revision, "--", filePath).Output()
	if err != nil {
		return nil, err
	}

	info, err := parseBlameOutput(out, commiter)
	if err != nil {
		return nil, err
	}
	if len(info) > 0 {
		return info, nil
	}
	author, commit, err := lastChangeForFile(repositoryPath, revision, filePath, commiter)
	if err != nil {
		return nil, err
	}
	if author != "" && commit != "" {
		info[author] = &authorInfo{commits: map[string]struct{}{commit: {}}}
	}
	return info, nil
}

func parseBlameHeader(line string) (string, bool) {
	fields := strings.Fields(line)
	if len(fields) < 3 {
		return "", false
	}
	hash := strings.TrimPrefix(fields[0], "^")
	if len(hash) < 7 {
		return "", false
	}
	if _, err := strconv.Atoi(fields[1]); err != nil {
		return "", false
	}
	if _, err := strconv.Atoi(fields[2]); err != nil {
		return "", false
	}
	return hash, true
}

func lastChangeForFile(repositoryPath string, revision string, filePath string, commiter bool) (string, string, error) {
	person := "%an"
	if commiter {
		person = "%cn"
	}
	out, err := exec.Command("git", "-C", repositoryPath, "log", "-1", "--format=%H%x00"+person, revision, "--", filePath).Output()
	if err != nil {
		return "", "", err
	}
	logLine := strings.TrimSpace(string(out))
	if logLine == "" {
		return "", "", nil
	}
	parts := strings.SplitN(logLine, "\x00", 2)
	if len(parts) != 2 {
		return "", "", nil
	}
	return parts[1], parts[0], nil
}

func parseBlameOutput(out []byte, commiter bool) (map[string]*authorInfo, error) {
	scanner := bufio.NewScanner(bytes.NewReader(out))
	info := make(map[string]*authorInfo)
	seenAuthors := make(map[string]string)
	var currentAuthor string
	var currentCommit string
	for scanner.Scan() {
		line := scanner.Text()
		if commit, ok := parseBlameHeader(line); ok {
			currentCommit = commit
			currentAuthor = seenAuthors[commit]
			continue
		}
		if !commiter && strings.HasPrefix(line, "author ") {
			currentAuthor = strings.TrimSpace(strings.TrimPrefix(line, "author "))
			if currentCommit != "" && currentAuthor != "" {
				seenAuthors[currentCommit] = currentAuthor
			}
			continue
		}
		if commiter && strings.HasPrefix(line, "committer ") {
			currentAuthor = strings.TrimSpace(strings.TrimPrefix(line, "committer "))
			if currentCommit != "" && currentAuthor != "" {
				seenAuthors[currentCommit] = currentAuthor
			}
			continue
		}
		if !strings.HasPrefix(line, "\t") || currentAuthor == "" || currentCommit == "" {
			continue
		}
		a := info[currentAuthor]
		if a == nil {
			a = &authorInfo{commits: map[string]struct{}{}}
			info[currentAuthor] = a
		}
		a.lines++
		a.commits[currentCommit] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return info, nil
}
