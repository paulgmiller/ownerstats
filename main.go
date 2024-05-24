package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type gopackage struct {
	Path        string
	LinesOfCode int64
}

func main() {
	// Define the root directory to start the walk from.
	root := os.Args[1]
	fmt.Printf("Walking %s\n", root)

	owners := map[string][]string{}
	counts := map[string]int64{}

	// Walk the filesystem starting from the root directory.
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error accessing path %q: %v\n", path, err)
			return err
		}

		// Check if the current file is named "owners.go".
		if info.IsDir() {
			return nil
		}

		dir := strings.TrimRight(path, info.Name())

		if strings.HasSuffix(info.Name(), ".go") {
			loc, err := countLines(path)
			if err != nil {
				fmt.Printf("couldn't count %s", path)
				return nil
			}
			counts[dir] += loc
			return nil
		}

		if info.Name() != "owners.txt" {
			return nil
		}

		people, err := getowners(path)
		if err != nil {
			fmt.Printf("couldn't parse owners %s", path)
			return nil
		}
		for _, owner := range people {
			owners[owner] = append(owners[owner], dir)
		}
		return nil
	})
	// Handle any errors encountered during the walk.
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", root, err)
		return
	}

	results := map[string][]gopackage{}
	for owner, pkgs := range owners {
		for _, pkg := range pkgs {
			results[owner] = append(results[owner], gopackage{Path: pkg, LinesOfCode: counts[pkg]})
		}
	}

	output, err := json.MarshalIndent(results, "", "    ")

	// Handle any errors encountered during the walk.
	if err != nil {
		fmt.Printf("can't serialize")
		return
	}
	fmt.Print(string(output))
}

func getowners(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return []string{}, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	var people []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Trim(line, " ")
		if strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			continue
		}
		if line == "" {
			continue
		}
		line = strings.TrimLeft(line, "*")
		line = strings.TrimLeft(line, "@")
		people = append(people, line)
	}

	if err := scanner.Err(); err != nil {
		return people, fmt.Errorf("error reading file: %w", err)
	}

	return people, nil
}

func countLines(filePath string) (int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var lineCount int64
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	return lineCount, nil
}
