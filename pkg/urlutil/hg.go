package urlutil

import "strings"

var (
	validHgPrefixes = []string{
		"hg::",
		"ssh://",
	}
)

// IsHgURL returns true if the provided str is a mercurial repository URL.
func IsHgURL(str string) bool {
	// Is there a common pattern we can search for?
	/*
		if IsURL(str) && strings.HasSuffix(str, ".git") {
			return true
		}
	*/
	for _, prefix := range validHgPrefixes {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	return false
}

// IsHgTransport returns true if the provided str is a mercurial transport by inspecting
// the prefix of the string for known protocols used in mercurial.
func IsHgTransport(str string) bool {
	return IsURL(str) || strings.HasPrefix(str, "hg::") || strings.HasPrefix(str, "hg@")
}

func CleanHgURL(str string) string {
	if strings.HasPrefix(str, "hg::") {
		return str[4:]
	}
	return str
}
