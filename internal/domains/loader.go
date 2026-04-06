package domains

import (
	"bufio"
	_ "embed"
	"strings"
)

//go:embed ai-domains.txt
var embeddedDomains string

func Load() ([]string, error) {
	var rules []string
	scanner := bufio.NewScanner(strings.NewReader(embeddedDomains))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		rules = append(rules, line)
	}
	return rules, scanner.Err()
}
