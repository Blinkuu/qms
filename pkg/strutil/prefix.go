package strutil

func WithPrefixOrDefault(prefix string, toPrefix string) string {
	if len(prefix) > 0 {
		return prefix + "." + toPrefix
	}

	return toPrefix
}
