package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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

func getLatestReleaseBinary() (io.Reader, error) {
	client := github.NewClient(nil)
	releases, _, err := client.Repositories.ListReleases(context.Background(), "hieutdle", "violet", &github.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch releases: %w", err)
	}

	// Get the latest release
	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found")
	}
	latestRelease := releases[0]
	assets := latestRelease.Assets

	// Determine the correct file based on the current OS and architecture
	platform := runtime.GOOS + "_"
	switch runtime.GOARCH {
	case "amd64":
		platform += "64-bit"
	case "arm64":
		platform += "arm64"
	default:
		return nil, fmt.Errorf("unsupported architecture: %s", runtime.GOARCH)
	}
	ext := ".tar.gz"

	var downloadURL string
	for _, asset := range assets {
		// Check if asset matches the platform and file extension
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
		return nil, fmt.Errorf("unable to download file: %w", err)
	}
	defer resp.Body.Close()

	// Read the entire content into memory
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	// Return the content as a bytes.Reader, which implements io.ReaderAt
	return bytes.NewReader(buf), nil
}

func applyUpdate(reader io.Reader) error {
	// Wrap the reader with gzip
	gzipReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("unable to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create a tar reader
	tarReader := tar.NewReader(gzipReader)

	// Iterate through the tar archive
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("error reading tarball: %w", err)
		}

		// Look for the binary file
		if header.Typeflag == tar.TypeReg && header.Name == "violet" { // Adjust file name if necessary
			// Apply the update
			if err := selfupdate.Apply(tarReader, selfupdate.Options{}); err != nil {
				return fmt.Errorf("error applying update: %w", err)
			}

			fmt.Println("Update applied successfully!")
			return nil
		}
	}

	return fmt.Errorf("binary file not found in tarball")
}
