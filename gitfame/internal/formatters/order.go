package formatters

import (
	"errors"
	"sort"
)

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
			if out[i].Files != out[j].Files {
				return out[i].Files > out[j].Files
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
			if out[i].Files != out[j].Files {
				return out[i].Files > out[j].Files
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
			if out[i].Commits != out[j].Commits {
				return out[i].Commits > out[j].Commits
			}
			return out[i].Name < out[j].Name
		})
	default:
		return nil, errors.New("invalid sort key")
	}

	return out, nil
}
