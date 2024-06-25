package autodetect

import (
	"os"
	"path/filepath"
)

type (
	Pattern       []string
	LanguePattern struct {
		Language string
		Patterns []Pattern
	}
)

var languagePatterns = []LanguePattern{
	{
		Language: "go",
		Patterns: []Pattern{
			{"go.mod", "go.sum"},
		},
	},
	{
		Language: "nodejs",
		Patterns: []Pattern{
			{"package.json", "package-lock.json"},
			{"yarn.lock"},
		},
	},
	{
		Language: "java",
		Patterns: []Pattern{
			{"pom.xml"},
			{"build.gradle"},
		},
	},
	{
		Language: "python",
		Patterns: []Pattern{
			{"setup.py", "requirements.txt"},
			{"Pipfile"},
			{"poetry.lock"},
		},
	},
	{
		Language: "ruby",
		Patterns: []Pattern{
			{"Gemfile", "Gemfile.lock"},
		},
	},
	{
		Language: "php",
		Patterns: []Pattern{
			{"composer.json", "composer.lock"},
		},
	},
	{
		Language: "rust",
		Patterns: []Pattern{
			{"Cargo.toml", "Cargo.lock"},
		},
	},
}

func PatternsByLanguage(workingdir string) map[string][]string {
	detectedPatterns := make(map[string][]string)

	for _, languagePattern := range languagePatterns {
		for _, pattern := range languagePattern.Patterns {
			allFilesExist := true
			for _, file := range pattern {
				if _, err := os.Stat(filepath.Join(workingdir, file)); os.IsNotExist(err) {
					allFilesExist = false
					break
				}
			}
			if allFilesExist {
				detectedPatterns[languagePattern.Language] = pattern
				break
			}
		}
	}

	return detectedPatterns
}
