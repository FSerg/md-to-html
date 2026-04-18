package converter

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/mozillazg/go-unidecode"
)

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func translitSlug(s string, used map[string]int) string {
	t := strings.ToLower(unidecode.Unidecode(s))
	t = slugRe.ReplaceAllString(t, "-")
	t = strings.Trim(t, "-")
	if t == "" {
		t = "section"
	}
	if n, ok := used[t]; ok && n > 0 {
		used[t] = n + 1
		return fmt.Sprintf("%s-%d", t, n)
	}
	used[t] = 1
	return t
}
