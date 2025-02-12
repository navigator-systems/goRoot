package ops

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ListScripts(dir string) ([]string, error) {
	var scripts []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Exclude hidden files and directories
		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		// Check if any part of the path is hidden
		parts := strings.Split(relPath, string(os.PathSeparator))
		for _, part := range parts {
			if strings.HasPrefix(part, ".") {
				return nil
			}
		}

		if !info.IsDir() {
			scripts = append(scripts, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return scripts, nil
}

func ReadScript(script, dir string) (string, string) {
	text := filepath.Join(dir, script)
	content, err := os.ReadFile(text)
	if err != nil {
		log.Println("Failed to read script file", err)
	}
	// Convert the path to lowercase
	lowerPath := strings.ToLower(script)
	path := strings.Replace(lowerPath, "/", "-", -1)
	path = strings.Replace(path, "_", "-", -1)

	return path, string(content)
}

func CreateFile(filename string) (*os.File, error) {
	// Create an uploads directory if it doesn’t exist
	if _, err := os.Stat("uploads"); os.IsNotExist(err) {
		os.Mkdir("uploads", 0755)
	}

	// Build the file path and create it
	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		return nil, err
	}

	return dst, nil
}
