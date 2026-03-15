package resources

import (
	"io/fs"
	"os"
	"path/filepath"
)

// ExtractTo extracts all embedded resources to the given directory.
// Overwrites unconditionally.
func ExtractTo(dir string) error {
	return fs.WalkDir(FS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dir, path)
		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}
		data, err := FS.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0644)
	})
}

// DiscoverSystemPromptPath returns the path to discover.md after extraction.
func DiscoverSystemPromptPath(resourceDir string) string {
	return filepath.Join(resourceDir, "system-prompts", "discover.md")
}

// PromptsDir returns the path to the prompts directory after extraction.
func PromptsDir(resourceDir string) string {
	return filepath.Join(resourceDir, "prompts")
}
