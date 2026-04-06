package gitfame

import (
	"errors"
	"fmt"
	"os"

	"github.com/Karambollla/gitfame/gitfame/internal/formatters"
	"github.com/Karambollla/gitfame/gitfame/internal/git"
	"github.com/Karambollla/gitfame/gitfame/internal/models"
	"github.com/Karambollla/gitfame/gitfame/internal/util"
)

// collects options for gitfame

type Options = models.Options

func Validate(o *Options) error {
	switch o.OrderBy {
	case "lines", "commits", "files", "":
	default:
		return errors.New("invalid --order-by, must be one of: lines,commits,files")
	}
	switch o.Format {
	case "tabular", "csv", "json", "json-lines", "":
	default:
		return errors.New("invalid --format, must be one of: tabular,csv,json,json-lines")
	}
	if o.Repository == "" {
		o.Repository = "."
	}

	if len(o.Languages) > 0 {
		langMap, err := formatters.LoadLanguageMap(o.Repository)
		if err != nil {
			fmt.Fprintln(os.Stderr, "warning: cannot read language_extensions.json:", err)
			return nil
		}
		o.LanguageMap = langMap
		o.Extensions = buildMergedExtensions(o)
	}

	return nil
}

func buildMergedExtensions(o *Options) []string {
	extSet := make(map[string]struct{})
	for _, e := range o.Extensions {
		ne := util.NormalizeExt(e)
		if ne != "" {
			extSet[ne] = struct{}{}
		}
	}
	for _, lang := range o.Languages {
		exts := formatters.ExtensionsForLanguage(lang, o.LanguageMap)
		if len(exts) == 0 {
			fmt.Fprintln(os.Stderr, "warning: unknown language:", lang)
			continue
		}
		for _, e := range exts {
			ne := util.NormalizeExt(e)
			if ne != "" {
				extSet[ne] = struct{}{}
			}
		}
	}
	merged := make([]string, 0, len(extSet))
	for e := range extSet {
		merged = append(merged, e)
	}
	return merged
}

func Run(o *Options) error {
	files, err := git.CollectFiles(o.Repository, o.Languages, o.Extensions, o.Exclude, o.RestrictTo, o.Revision, o.LanguageMap)
	if err != nil {
		return err
	}

	res, err := git.CollectBlame(o.Repository, o.Revision, files, o.UseCommitter)
	if err != nil {
		return err
	}

	sorted, err := formatters.OrderByLines(res, o.OrderBy)
	if err != nil {
		return err
	}

	out := formatters.Format(sorted, o.Format)
	fmt.Print(out)
	return nil
}
