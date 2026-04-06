package formatters

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/Karambollla/gitfame/gitfame/internal/models"
)

type BlameLine = models.BlameLine

func Format(blame []BlameLine, format string) string {
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
		b, err := json.Marshal(blame)
		if err != nil {
			return ""
		}
		return string(b)

	case "json-lines":
		lines := blame
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
	}
	return ""
}
