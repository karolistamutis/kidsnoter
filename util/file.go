package util

import (
	"fmt"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const layout = "2006-01-02T15:04:05.000000Z"

func ExtractExtensionFromURL(rawURL string) (string, error) {
	// Remove any surrounding quotes
	cleanURL := strings.Trim(rawURL, "\"")

	// Parse the URL
	parsedURL, err := url.Parse(cleanURL)
	if err != nil {
		return "", err
	}

	// Get the file name from the URL path
	fileName := path.Base(parsedURL.Path)

	// Split the file name by dots
	parts := strings.Split(fileName, ".")

	// Check if there's an extension
	if len(parts) > 1 {
		// Return the last part (extension)
		return parts[len(parts)-1], nil
	}

	// If no extension found, return an empty string
	return "", nil
}

func GenerateFolderName(childName string, date string, albumID int, albumTitle string) (string, error) {
	formattedDate, err := formatDate(date)
	if err != nil {
		// Fallback to using unformatted date
		formattedDate = date
	}

	normalizedChildName, err := normalize(childName)
	if err != nil {
		return "", fmt.Errorf("failed to normalize the child name for album directory creation (%s): %v", childName, err)
	}

	normalizedTitle, err := normalize(albumTitle)
	if err != nil {
		return "", fmt.Errorf("failed to normalize the title for album directory creation (%s): %v", albumTitle, err)
	}

	// childName/YYYY/MM_albumID_title
	return fmt.Sprintf("%s/%s_%d_%s", normalizedChildName, formattedDate, albumID, normalizedTitle), nil
}

func ExpandTilde(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(homeDir, path[2:]), nil
	}
	return path, nil
}

// FileExistsAndMatches checks if a file exists and matches a given size.
func FileExistsAndMatches(filepath string, expectedSize int) bool {
	if info, err := os.Stat(filepath); err == nil {
		return info.Size() == int64(expectedSize)
	}
	return false
}

func formatDate(date string) (string, error) {
	t, err := time.Parse(layout, date)
	if err != nil {
		return "", fmt.Errorf("failed to parse date: %v", err)
	}

	// Format the time for use in directory structure creation - YYYY/MM
	return t.Format("2006/01"), nil
}

func normalize(input string) (string, error) {
	// Create a transformer chain to remove diacritics and normalize the string
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

	clean, _, err := transform.String(t, input)
	if err != nil {
		return "", err
	}

	// Remove all non-alphanumeric characters except spaces
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s]+`)
	clean = reg.ReplaceAllString(clean, "")

	// Replace spaces with a single underscore and trim
	clean = strings.Join(strings.Fields(clean), "_")

	return clean, nil
}
