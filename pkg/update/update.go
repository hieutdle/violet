package update

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/google/go-github/v48/github"
	"github.com/minio/selfupdate"
)

func Update() error {
	file, err := getLatestReleaseBinary()
	if err != nil {
		return fmt.Errorf("failed to get latest release binary: %w", err)
	}

	// Apply the update (unzip and pass to Apply function)
	if err := applyUpdate(file); err != nil {
		return fmt.Errorf("failed to apply update: %w", err)
	}

	return nil
}

// getLatestReleaseBinary fetches the latest release binary for the current OS and architecture
func getLatestReleaseBinary() (io.Reader, error) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "hieutdle", "violet", &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch releases: %v", err)
	}

	// Get the latest release
	latestRelease := releases[0]
	assets := latestRelease.Assets

	// Determine the correct file based on the current OS and architecture
	ext := ".tar.gz"
	platform := runtime.GOOS + "_" + runtime.GOARCH
	var downloadURL string
	for _, asset := range assets {
		// Check if asset matches the platform
		if strings.Contains(asset.GetName(), platform) && strings.HasSuffix(asset.GetName(), ext) {
			downloadURL = asset.GetBrowserDownloadURL()
			break
		}
	}

	if downloadURL == "" {
		return nil, fmt.Errorf("no compatible release found for %s", platform)
	}

	// Download the file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return nil, fmt.Errorf("unable to download file: %v", err)
	}
	defer resp.Body.Close()

	// Read the entire content into memory
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %v", err)
	}

	// Return the content as a bytes.Reader, which implements io.ReaderAt
	return bytes.NewReader(buf), nil
}

// applyUpdate applies the update using the minio selfupdate package
func applyUpdate(reader io.Reader) error {
	// Use the bytes.Reader (which implements io.ReaderAt) to work with zip
	zipReader, err := zip.NewReader(reader.(io.ReaderAt), int64(reader.(*bytes.Reader).Len()))
	if err != nil {
		return fmt.Errorf("unable to read zip file: %v", err)
	}

	// Get the first file from the zip archive
	for _, file := range zipReader.File {
		if file.Name == "violet" {
			f, err := file.Open()
			if err != nil {
				return fmt.Errorf("unable to open file in zip: %v", err)
			}
			defer f.Close()

			// Use the file stream to apply the update
			if err := selfupdate.Apply(f, selfupdate.Options{}); err != nil {
				return fmt.Errorf("error applying update: %v", err)
			}

			// Apply the update (use the reader passed)
			fmt.Println("Update applied successfully!")
			return nil
		}
	}

	return fmt.Errorf("binary file not found in archive")
}
