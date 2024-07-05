package util

import (
	"bufio"
	"fmt"
	"strings"
)

// ParseFrontmatter returns title and date strings for the album given input string containing frontmatter
func ParseFrontmatter(content string) (string, string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	var title string
	var date string

	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "---") {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.Trim(strings.TrimSpace(parts[1]), "\"")
				if key == "title" {
					title = value
				} else if key == "date" {
					date = value
				}
			}
		}
	}

	if title == "" || date == "" {
		return "", "", fmt.Errorf("invalid frontmatter")
	}

	return title, date, nil
}
