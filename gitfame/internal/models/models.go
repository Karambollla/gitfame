package models

type BlameLine struct {
	Name    string
	Lines   int
	Commits int
	Files   int
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
