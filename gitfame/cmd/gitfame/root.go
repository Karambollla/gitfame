package main

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/slon/shad-go/gitfame/internal/gitfame"
	"gitlab.com/slon/shad-go/gitfame/internal/models"
	"gitlab.com/slon/shad-go/gitfame/internal/util"
)

var (
	repository   string
	revision     string
	orderBy      string
	useCommitter bool
	format       string

	extensions string
	languages  string
	exclude    string
	restrictTo string

	showProgress bool
)

var rootCmd = &cobra.Command{
	Use:   "gitfame",
	Short: "show stats of git",
	Run: func(cmd *cobra.Command, args []string) {
		opts := models.Options{
			Repository:   repository,
			Revision:     revision,
			OrderBy:      orderBy,
			UseCommitter: useCommitter,
			Format:       format,
			Extensions:   util.SplitCSV(extensions),
			Languages:    util.SplitCSV(languages),
			Exclude:      util.SplitCSV(exclude),
			RestrictTo:   util.SplitCSV(restrictTo),
			ShowProgress: showProgress,
		}
		if err := gitfame.Validate(&opts); err != nil {
			slog.Error("failed validating options", "error", err)
			os.Exit(2)
		}
		if err := gitfame.Run(&opts); err != nil {
			slog.Error("failed running options", "error", err)
			os.Exit(1)
		}

	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		slog.Error("failed executing", "error", err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVar(&repository, "repository", ".", "path to Git repository; default current directory")
	rootCmd.PersistentFlags().StringVar(&revision, "revision", "HEAD", "revision to analyze; default HEAD")
	rootCmd.PersistentFlags().StringVar(&orderBy, "order-by", "lines", "order by one of: lines,commits,files")
	rootCmd.PersistentFlags().BoolVar(&useCommitter, "use-committer", false, "use committer instead of author")
	rootCmd.PersistentFlags().StringVar(&format, "format", "tabular", "output format: tabular,csv,json,json-lines")

	rootCmd.PersistentFlags().StringVar(&extensions, "extensions", "", "comma-separated list of extensions to include, e.g. '.go,.md'")
	rootCmd.PersistentFlags().StringVar(&languages, "languages", "", "comma-separated list of languages to include, e.g. 'go,markdown'")
	rootCmd.PersistentFlags().StringVar(&exclude, "exclude", "", "comma-separated glob patterns to exclude, e.g. 'foo/*,bar/*'")
	rootCmd.PersistentFlags().StringVar(&restrictTo, "restrict-to", "", "comma-separated glob patterns to restrict to")

	rootCmd.PersistentFlags().BoolVar(&showProgress, "progress", false, "show progress to stderr")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

//bozhe hrani golang
