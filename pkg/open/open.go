package open

import (
	"fmt"
	"os/exec"
	"runtime"
)

var websiteMap = map[string][]string{
	"https://www.google.com":       {"go", "gg", "google"},
	"https://www.github.com":       {"git", "gh", "github"},
	"https://translate.google.com": {"trans", "translate"},
	"https://chat.openai.com/":     {"chat", "gpt", "chatgpt"},
	"https://www.facebook.com":     {"fb", "face", "facebook"},
}

var reverseWebsiteMap = map[string]string{}

func init() {
	for url, ids := range websiteMap {
		for _, id := range ids {
			reverseWebsiteMap[id] = url
		}
	}
}

func OpenWebsite(id string) error {
	url, exists := reverseWebsiteMap[id]
	if !exists {
		return fmt.Errorf("invalid website ID: %s", id)
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open website: %v", err)
	}
	return nil
}
