package autodetect

import (
	"os"
	"path/filepath"
	"testing"

	"gotest.tools/v3/assert"
	"gotest.tools/v3/env"
	"gotest.tools/v3/fs"
)

func TestWriteFile(t *testing.T) {
	tests := []struct {
		name     string
		language string
		files    []string
	}{
		{name: "golang", language: "go", files: []string{"go.mod", "go.sum"}},
		{name: "nodejs-npm", language: "nodejs", files: []string{"package.json", "package-lock.json"}},
		{name: "nodejs-yarn", language: "nodejs", files: []string{"yarn.lock"}},
		{name: "java-maven", language: "java", files: []string{"pom.xml"}},
		{name: "java-gradle", language: "java", files: []string{"build.gradle"}},
		{name: "python-setup", language: "python", files: []string{"setup.py", "requirements.txt"}},
		{name: "python-pipfile", language: "python", files: []string{"Pipfile"}},
		{name: "python-poetry", language: "python", files: []string{"poetry.lock"}},
		{name: "ruby", language: "ruby", files: []string{"Gemfile", "Gemfile.lock"}},
		{name: "php", language: "php", files: []string{"composer.json", "composer.lock"}},
		{name: "rust", language: "rust", files: []string{"Cargo.toml", "Cargo.lock"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpdir := fs.NewDir(t, t.Name())
			defer tmpdir.Remove()

			defer env.ChangeWorkingDir(t, tmpdir.Path())()

			for _, file := range tt.files {
				err := os.WriteFile(filepath.Join(tmpdir.Path(), file), []byte("random content"), 0o644)
				assert.NilError(t, err)
			}

			patterns := PatternsByLanguage(tmpdir.Path())
			assert.DeepEqual(t, patterns, map[string][]string{
				tt.language: tt.files,
			})
		})
	}
}
