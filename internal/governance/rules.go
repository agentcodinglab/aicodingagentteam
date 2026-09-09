// Package governance implements rule checkers for governance violations.
package governance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// checkEmoji detects emoji characters in code (UI icon rule).
var emojiRe = regexp.MustCompile(`[\x{1F300}-\x{1F9FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]|[\x{FE0F}]|\\u\{[0-9a-fA-F]{4,5}\}`)

func checkEmoji(path, content string) []Violation {
	if !strings.HasSuffix(path, ".tsx") && !strings.HasSuffix(path, ".jsx") && !strings.HasSuffix(path, ".vue") {
		return nil
	}
	var out []Violation
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if emojiRe.MatchString(line) {
			// Skip lines that are clearly comments about emoji
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
				continue
			}
			out = append(out, Violation{
				RuleID:   "ui-emoji-icon",
				Severity: "advisory",
				Path:     path,
				Detail:   "emoji in code at line " + itoa(i+1),
			})
		}
	}
	return out
}

// checkHardcodedColor detects hardcoded hex colors in UI files.
var hexColorRe = regexp.MustCompile(`#[0-9a-fA-F]{3,8}`)

func checkHardcodedColor(path, content string) []Violation {
	if !strings.HasSuffix(path, ".tsx") && !strings.HasSuffix(path, ".jsx") &&
		!strings.HasSuffix(path, ".css") && !strings.HasSuffix(path, ".scss") &&
		!strings.HasSuffix(path, ".vue") {
		return nil
	}
	var out []Violation
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if hexColorRe.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") {
				continue
			}
			// Allow CSS custom property definitions
			if strings.Contains(line, "--") && strings.Contains(line, ":") {
				continue
			}
			out = append(out, Violation{
				RuleID:   "ui-hardcoded-color",
				Severity: "advisory",
				Path:     path,
				Detail:   "hardcoded color at line " + itoa(i+1),
			})
		}
	}
	return out
}

// checkSecretLeak detects common secret/credential patterns.
var secretPatterns = []struct {
	name    string
	pattern *regexp.Regexp
}{
	{"AWS_ACCESS_KEY", regexp.MustCompile(`AKIA[0-9A-Z]{16}`)},
	{"AWS_SECRET_KEY", regexp.MustCompile(`(?i)aws.{0,20}secret.{0,20}['"][0-9a-zA-Z/+=]{40}['"]`)},
	{"GITHUB_TOKEN", regexp.MustCompile(`ghp_[0-9a-zA-Z]{20,}`)},
	{"GITHUB_OAUTH", regexp.MustCompile(`gho_[0-9a-zA-Z]{36}`)},
	{"SLACK_TOKEN", regexp.MustCompile(`xox[bpoas]-[0-9a-zA-Z-]+`)},
	{"PRIVATE_KEY", regexp.MustCompile(`-----BEGIN (RSA |EC |DSA |OPENSSH )?PRIVATE KEY-----`)},
	{"GENERIC_API_KEY", regexp.MustCompile(`(?i)(api[_-]?key|api[_-]?secret|auth[_-]?token)\s*[:=]\s*['"][0-9a-zA-Z_\-]{20,}['"]`)},
	{"SK_KEY", regexp.MustCompile(`sk-[0-9a-zA-Z]{20,}`)},
	{"BEARER_TOKEN", regexp.MustCompile(`(?i)bearer\s+[0-9a-zA-Z_\-\.=]{20,}`)},
}

func checkSecretLeak(path, content string) []Violation {
	// Skip test files and example files
	lower := strings.ToLower(path)
	if strings.Contains(lower, "_test.") || strings.Contains(lower, ".example") || strings.Contains(lower, ".sample") {
		return nil
	}
	var out []Violation
	for _, sp := range secretPatterns {
		if sp.pattern.MatchString(content) {
			out = append(out, Violation{
				RuleID:   "sec-secret-leak",
				Severity: "blocking",
				Path:     path,
				Detail:   "potential secret leak: " + sp.name,
			})
		}
	}
	return out
}

// checkSQLInjection detects string-concatenated SQL queries.
var sqlConcatRe = regexp.MustCompile(`(?i)(SELECT|INSERT|UPDATE|DELETE|DROP)\s+.*\+`)

func checkSQLInjection(path, content string) []Violation {
	if !strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".java") &&
		!strings.HasSuffix(path, ".py") && !strings.HasSuffix(path, ".ts") &&
		!strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".rb") {
		return nil
	}
	var out []Violation
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if sqlConcatRe.MatchString(line) {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") {
				continue
			}
			out = append(out, Violation{
				RuleID:   "sec-sql-injection",
				Severity: "blocking",
				Path:     path,
				Detail:   "potential SQL injection at line " + itoa(i+1),
			})
		}
	}
	return out
}

// callSiteRe matches fetch("...") and axios("...") and axios.get("...") etc.
var callSiteRe = regexp.MustCompile(`(?:fetch|axios(?:\.\w+)?)\s*\(\s*['"]([^'"]+)['"]`)

// extractPathFromURL strips query strings and base URL prefixes from an API path.
func extractPathFromURL(raw string) string {
	if idx := strings.Index(raw, "?"); idx >= 0 {
		raw = raw[:idx]
	}
	if idx := strings.Index(raw, "://"); idx >= 0 {
		rest := raw[idx+3:]
		if slash := strings.Index(rest, "/"); slash >= 0 {
			raw = rest[slash:]
		} else {
			return ""
		}
	}
	return raw
}

// openapiSpec is a minimal OpenAPI JSON structure for path extraction.
type openapiSpec struct {
	Paths map[string]json.RawMessage `json:"paths"`
}

// loadOpenAPIPaths reads an OpenAPI/JSON file and extracts declared path patterns.
// It first tries encoding/json parsing; if that fails it falls back to line scanning.
func loadOpenAPIPaths(data []byte) map[string]bool {
	paths := make(map[string]bool)

	// Try JSON parsing first
	var spec openapiSpec
	if err := json.Unmarshal(data, &spec); err == nil && len(spec.Paths) > 0 {
		for p := range spec.Paths {
			if strings.HasPrefix(p, "/") {
				paths[p] = true
			}
		}
		return paths
	}

	// Fallback: line-by-line scanning for path patterns
	content := string(data)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "\"/") {
			end := strings.Index(trimmed[1:], "\"")
			if end > 0 {
				p := trimmed[1 : 1+end]
				if strings.HasPrefix(p, "/") {
					paths[p] = true
				}
			}
		}
	}
	return paths
}

// checkAPIContract detects fetch/axios calls that don't match OpenAPI paths.
// It scans for a co-located openapi.* spec and checks that all frontend
// call sites reference paths declared in the spec.
func checkAPIContract(path, content string) []Violation {
	if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") &&
		!strings.HasSuffix(path, ".js") && !strings.HasSuffix(path, ".jsx") {
		return nil
	}
	if !strings.Contains(content, "fetch(") && !strings.Contains(content, "axios") {
		return nil
	}

	dir := filepath.Dir(path)
	specPaths := []string{
		filepath.Join(dir, "openapi.json"),
		filepath.Join(dir, "openapi.yaml"),
		filepath.Join(dir, "openapi.yml"),
		filepath.Join(dir, "..", "openapi.json"),
		filepath.Join(dir, "..", "openapi.yaml"),
	}
	var contractPaths map[string]bool
	for _, sp := range specPaths {
		if data, err := os.ReadFile(sp); err == nil {
			contractPaths = loadOpenAPIPaths(data)
			if len(contractPaths) > 0 {
				break
			}
		}
	}

	var out []Violation
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		matches := callSiteRe.FindAllStringSubmatch(line, -1)
		for _, m := range matches {
			if len(m) < 2 {
				continue
			}
			rawPath := m[1]
			apiPath := extractPathFromURL(rawPath)
			if apiPath == "" || !strings.HasPrefix(apiPath, "/") {
				continue
			}
			if len(contractPaths) > 0 {
				if !contractPaths[apiPath] {
					matched := false
					for p := range contractPaths {
						if strings.Contains(p, "{") && pathPrefixMatch(apiPath, p) {
							matched = true
							break
						}
					}
					if !matched {
						out = append(out, Violation{
							RuleID:   "api-contract-mismatch",
							Severity: "blocking",
							Path:     path,
							Detail:   fmt.Sprintf("API call at line %d references %q not in OpenAPI spec", i+1, apiPath),
						})
					}
				}
			}
		}
	}
	return out
}

// pathPrefixMatch checks if a concrete path matches a pattern with {param} segments.
func pathPrefixMatch(concrete, pattern string) bool {
	concreteParts := strings.Split(concrete, "/")
	patternParts := strings.Split(pattern, "/")
	if len(concreteParts) != len(patternParts) {
		return false
	}
	for i := 0; i < len(concreteParts); i++ {
		pp := patternParts[i]
		if strings.HasPrefix(pp, "{") && strings.HasSuffix(pp, "}") {
			continue
		}
		if concreteParts[i] != pp {
			return false
		}
	}
	return true
}

// checkTodoPlaceholder detects TODO/FIXME/HACK placeholders in code.
var todoRe = regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|XXX|TEMP)\b`)

func checkTodoPlaceholder(path, content string) []Violation {
	var out []Violation
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if todoRe.MatchString(line) {
			out = append(out, Violation{
				RuleID:   "eng-todo-placeholder",
				Severity: "advisory",
				Path:     path,
				Detail:   "placeholder at line " + itoa(i+1),
			})
		}
	}
	return out
}

// checkFakeData detects hardcoded test/fake data in non-test files.
var fakeDataPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(test@example\.com|foo@bar\.com|user@test)\b`),
	regexp.MustCompile(`(?i)\b(lorem ipsum|dolor sit amet)\b`),
	regexp.MustCompile(`(?i)\b(fake|dummy|mock|stub)\s+(data|user|email|name)\b`),
}

func checkFakeData(path, content string) []Violation {
	lower := strings.ToLower(path)
	if strings.Contains(lower, "_test.") || strings.Contains(lower, ".test.") ||
		strings.Contains(lower, "testdata") || strings.Contains(lower, "fixture") {
		return nil
	}
	var out []Violation
	for _, re := range fakeDataPatterns {
		if re.MatchString(content) {
			out = append(out, Violation{
				RuleID:   "eng-fake-data",
				Severity: "advisory",
				Path:     path,
				Detail:   "fake/test data detected",
			})
			break // one finding per rule is enough
		}
	}
	return out
}

// itoa is a simple int-to-string helper (avoids strconv import in this file).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
