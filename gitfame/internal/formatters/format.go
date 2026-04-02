package formatters

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"gitlab.com/slon/shad-go/gitfame/internal/models"
)

type BlameLine = models.BlameLine

func Format(blame []BlameLine, format string) string {
	type outLine struct {
		Name    string `json:"name"`
		Lines   int    `json:"lines"`
		Commits int    `json:"commits"`
		Files   int    `json:"files"`
	}

	toOut := func(src []BlameLine) []outLine {
		out := make([]outLine, 0, len(src))
		for _, line := range src {
			out = append(out, outLine{
				Name:    line.Name,
				Lines:   line.Lines,
				Commits: line.Commits,
				Files:   line.Files,
			})
		}
		return out
	}

	switch format {
	case "", "tabular":
		var buf bytes.Buffer
		tw := tabwriter.NewWriter(&buf, 0, 0, 1, ' ', 0)
		_, _ = fmt.Fprintln(tw, "Name\tLines\tCommits\tFiles")
		for _, line := range blame {
			_, _ = fmt.Fprintf(tw, "%s\t%d\t%d\t%d\n", line.Name, line.Lines, line.Commits, line.Files)
		}
		_ = tw.Flush()
		return buf.String()

	case "csv":
		var buf bytes.Buffer
		w := csv.NewWriter(&buf)
		_ = w.Write([]string{"Name", "Lines", "Commits", "Files"})
		for _, line := range blame {
			_ = w.Write([]string{
				line.Name,
				strconv.Itoa(line.Lines),
				strconv.Itoa(line.Commits),
				strconv.Itoa(line.Files),
			})
		}
		w.Flush()
		return buf.String()

	case "json":
		b, err := json.Marshal(toOut(blame))
		if err != nil {
			return ""
		}
		return string(b)

	case "json-lines":
		lines := toOut(blame)
		var buf bytes.Buffer
		for i, line := range lines {
			b, err := json.Marshal(line)
			if err != nil {
				continue
			}
			buf.Write(b)
			if i != len(lines)-1 {
				buf.WriteByte('\n')
			}
		}
		return buf.String()

	default:
		return Format(blame, "tabular")
	}
}

func LoadLanguageMap(repositoryPath string) (*map[string][]string, error) {
	cfgPath := filepath.Join(repositoryPath, "configs", "language_extensions.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}

	var langMap map[string][]string
	if err := json.Unmarshal(data, &langMap); err != nil {
		return nil, err
	}

	norm := make(map[string][]string, len(langMap))
	for lang, exts := range langMap {
		key := strings.ToLower(strings.TrimSpace(lang))
		if key == "" {
			continue
		}

		out := make([]string, 0, len(exts))
		//nolint:staticcheck
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
		// nolint:staticcheck
		norm[key] = out
	}

	return &norm, nil
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
	for _, e := range exts {
		out = append(out, e)
	}

	return out
}

func SplitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func OrderByLines(blame []BlameLine, sortby string) ([]BlameLine, error) {
	type agg struct {
		lines   int
		commits int
		files   int
	}
	m := make(map[string]*agg)
	for _, b := range blame {
		a, ok := m[b.Name]
		if !ok {
			a = &agg{}
			m[b.Name] = a
		}
		a.lines += b.Lines
		a.commits += b.Commits
		a.files += b.Files
	}

	out := make([]BlameLine, 0, len(m))
	for name, a := range m {
		out = append(out, BlameLine{
			Name:    name,
			Lines:   a.lines,
			Commits: a.commits,
			Files:   a.files,
		})
	}

	switch sortby {
	case "", "lines":
		sort.Slice(out, func(i, j int) bool {
			if out[i].Lines != out[j].Lines {
				return out[i].Lines > out[j].Lines
			}
			if out[i].Commits != out[j].Commits {
				return out[i].Commits > out[j].Commits
			}
			return out[i].Name < out[j].Name
		})
	case "commits":
		sort.Slice(out, func(i, j int) bool {
			if out[i].Commits != out[j].Commits {
				return out[i].Commits > out[j].Commits
			}
			if out[i].Lines != out[j].Lines {
				return out[i].Lines > out[j].Lines
			}
			return out[i].Name < out[j].Name
		})
	case "files":
		sort.Slice(out, func(i, j int) bool {
			if out[i].Files != out[j].Files {
				return out[i].Files > out[j].Files
			}
			if out[i].Lines != out[j].Lines {
				return out[i].Lines > out[j].Lines
			}
			return out[i].Name < out[j].Name
		})
	default:
		return nil, fmt.Errorf("invalid sort key: %s", sortby)
	}

	return out, nil

}
