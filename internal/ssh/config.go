package appssh

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// sshHostConfig holds resolved values from ~/.ssh/config for a single host alias.
type sshHostConfig struct {
	HostName      string
	Port          string
	User          string
	IdentityFiles []string
}

// resolveSSHConfig reads ~/.ssh/config and returns merged config for the given
// host alias.  Only a useful subset of keywords is supported: Host, HostName,
// Port, User, IdentityFile.  Wildcard Host patterns are matched but specific
// matches take priority (openssh "first match wins" semantics per keyword).
func resolveSSHConfig(alias string) *sshHostConfig {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	path := filepath.Join(home, ".ssh", "config")
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var result sshHostConfig
	inMatchingBlock := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// skip comments and blank lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val := splitConfigLine(line)
		if key == "" {
			continue
		}

		lower := strings.ToLower(key)

		if lower == "host" {
			// New Host block — check if any pattern matches alias
			inMatchingBlock = hostPatternMatch(val, alias)
			continue
		}

		if !inMatchingBlock {
			continue
		}

		// First-match-wins per keyword (like openssh)
		switch lower {
		case "hostname":
			if result.HostName == "" {
				result.HostName = val
			}
		case "port":
			if result.Port == "" {
				if n, err := strconv.Atoi(val); err == nil && n > 0 && n <= 65535 {
					result.Port = val
				}
			}
		case "user":
			if result.User == "" {
				result.User = val
			}
		case "identityfile":
			// Collect all IdentityFile values (openssh supports multiple)
			result.IdentityFiles = append(result.IdentityFiles, expandTilde(val, home))
		}
	}

	if result.HostName == "" && result.Port == "" && result.User == "" && len(result.IdentityFiles) == 0 {
		return nil
	}
	return &result
}

// splitConfigLine splits "Key Value" or "Key=Value" into (key, value).
func splitConfigLine(line string) (string, string) {
	// Handle Key=Value
	if idx := strings.IndexByte(line, '='); idx >= 0 {
		return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:])
	}
	// Handle Key Value (split on first whitespace)
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		parts = strings.SplitN(line, "\t", 2)
	}
	if len(parts) < 2 {
		return parts[0], ""
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

// hostPatternMatch checks if alias matches any pattern in the space-separated
// Host value.  Supports "*" and "?" globs.
func hostPatternMatch(patterns, alias string) bool {
	for _, p := range strings.Fields(patterns) {
		if p == "*" {
			return true
		}
		matched, _ := filepath.Match(p, alias)
		if matched {
			return true
		}
	}
	return false
}

// expandTilde replaces leading ~ with home directory.
func expandTilde(path, home string) string {
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	if path == "~" {
		return home
	}
	return path
}
