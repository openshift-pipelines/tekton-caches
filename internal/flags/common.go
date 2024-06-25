package flags

import (
	"fmt"
	"os"

	"github.com/openshift-pipelines/tekton-caches/internal/autodetect"
	"github.com/spf13/cobra"
)

var PatternsFlag = "pattern"

func Patterns(cmd *cobra.Command, workingdir string) ([]string, error) {
	patterns, err := cmd.Flags().GetStringArray(PatternsFlag)
	if err != nil {
		return []string{}, err
	}
	if len(patterns) == 0 {
		// NOTE(chmouel): on multiples languages we use a single cache target, it
		// ust make things simpler
		// on very large monorepo this might be a problem
		languages := autodetect.PatternsByLanguage(workingdir)
		if len(languages) == 0 {
			return []string{}, fmt.Errorf("didn't detect any language, please specify the patterns with --%s flag", PatternsFlag)
		}
		for language, files := range languages {
			fmt.Fprintf(os.Stderr, "Detected project language %s\n", language)
			for _, file := range files {
				// NOTE(chmouel): we are using a glob pattern to match the top dir not the subdirs
				// but that's fine since most of the time most people don't use
				// composed dependencies workspaces (except the rustaceans)
				patterns = append(patterns, fmt.Sprintf("*%s", file))
			}
		}
	}
	return patterns, nil
}
