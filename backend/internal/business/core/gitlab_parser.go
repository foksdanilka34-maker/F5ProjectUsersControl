package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	numericBranchRe = regexp.MustCompile(`(?:^|/)(\d{1,9})[-_]`)
	nonSlugRe       = regexp.MustCompile(`[^a-z0-9]+`)
)

var translitTable = map[rune]string{
	'а': "a", 'б': "b", 'в': "v", 'г': "g", 'д': "d", 'е': "e", 'ё': "e", 'ж': "zh",
	'з': "z", 'и': "i", 'й': "y", 'к': "k", 'л': "l", 'м': "m", 'н': "n", 'о': "o",
	'п': "p", 'р': "r", 'с': "s", 'т': "t", 'у': "u", 'ф': "f", 'х': "h", 'ц': "c",
	'ч': "ch", 'ш': "sh", 'щ': "sch", 'ъ': "", 'ы': "y", 'ь': "", 'э': "e", 'ю': "yu",
	'я': "ya",
}

func NormalizeTaskKeyPrefix(prefix string) string {
	prefix = strings.TrimSpace(prefix)
	if prefix == "" {
		return "F5"
	}
	return strings.ToUpper(prefix)
}

func BuildTaskKey(prefix string, taskID int64) string {
	return fmt.Sprintf("%s-%d", NormalizeTaskKeyPrefix(prefix), taskID)
}

func ParseTaskID(prefix string, sources ...string) int64 {
	keyRe, err := regexp.Compile(`(?i)\b` + regexp.QuoteMeta(NormalizeTaskKeyPrefix(prefix)) + `[-_ ]?(\d{1,9})\b`)
	if err != nil {
		return 0
	}

	for _, source := range sources {
		if source == "" {
			continue
		}
		if m := keyRe.FindStringSubmatch(source); len(m) == 2 {
			if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
				return id
			}
		}
	}

	for _, source := range sources {
		if m := numericBranchRe.FindStringSubmatch(source); len(m) == 2 {
			if id, err := strconv.ParseInt(m[1], 10, 64); err == nil {
				return id
			}
		}
	}

	return 0
}

func BuildBranchName(prefix string, taskID int64, title string) string {
	slug := Slugify(title)
	base := strings.ToLower(BuildTaskKey(prefix, taskID))
	if slug == "" {
		return "feature/" + base
	}
	if len(slug) > 40 {
		slug = strings.Trim(slug[:40], "-")
	}
	return "feature/" + base + "-" + slug
}

func Slugify(value string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(value) {
		if replacement, ok := translitTable[r]; ok {
			sb.WriteString(replacement)
			continue
		}
		sb.WriteRune(r)
	}
	return strings.Trim(nonSlugRe.ReplaceAllString(sb.String(), "-"), "-")
}
