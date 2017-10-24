// Copyright (c) 2017 AnUnnamedProject
// Distributed under the MIT software license, see the accompanying
// file LICENSE or http://www.opensource.org/licenses/mit-license.php.

package framework

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getAbsolutePath check if path exists and returns an absolute path.
func getAbsolutePath(path string) (string, error) {
	// Check if path exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	// Check if the directory is valid
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}

	// If path is already absolute, return it.
	if filepath.IsAbs(path) {
		return path, nil
	}

	// Get current absolute path
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Join absolute path with provided path
	absoluteDir := filepath.Join(wd, path)
	_, err = os.Stat(absoluteDir)
	if err != nil {
		return "", err
	}

	return absoluteDir, nil
}

// lookupFile check if file exists and returns absolute path, os.FileInfo and error.
func lookupFile(path string) (string, os.FileInfo, error) {
	requestPath := filepath.ToSlash(filepath.Clean(path))
	if !strings.Contains(requestPath, "/") {
		return "", nil, nil
	}
	if len(requestPath) > 1 && requestPath[0] != '/' {
		return "", nil, nil
	}

	filePath := filepath.Join(App.staticDir, requestPath[1:])
	if fi, _ := os.Stat(filePath); fi != nil {
		return filePath, fi, nil
	}

	return "", nil, nil
}
