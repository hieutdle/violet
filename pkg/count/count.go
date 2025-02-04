package count

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
)

func CountGitLines() int {
	cmd := exec.Command("git", "ls-files")
	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to run git ls-files: %v", err)
	}

	files := bytes.Split(bytes.TrimSpace(output), []byte("\n"))
	var totalLines int

	for _, file := range files {
		f, err := os.Open(string(file))
		if err != nil {
			log.Printf("Skipping file %s: %v", file, err)
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			totalLines++
		}
		f.Close()
	}

	return totalLines
}
