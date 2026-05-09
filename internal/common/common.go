/*
Copyright © 2026 Two Tech Studio
*/
package common

// IsHex checks if a string is a valid hexadecimal string.
func IsHex(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		if !((c >= '0' && c <= '9') ||
			(c >= 'a' && c <= 'f') ||
			(c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}

// ShortHash returns the first 7 characters of a hash, or the full string if shorter.
func ShortHash(s string) string {
	if len(s) <= 7 {
		return s
	}
	return s[:7]
}
