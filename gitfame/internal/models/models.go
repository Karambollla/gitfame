package models

type BlameLine struct {
	Name    string `json:"name"`
	Lines   int    `json:"lines"`
	Commits int    `json:"commits"`
	Files   int    `json:"files"`
}

type Options struct {
	Repository   string
	Revision     string
	OrderBy      string
	UseCommitter bool
	Format       string
	Extensions   []string
	Languages    []string
	Exclude      []string
	RestrictTo   []string
	ShowProgress bool
	LanguageMap  *map[string][]string
}
