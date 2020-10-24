package utils

// IsStringInSlice check if a string is in a given slice
func IsStringInSlice(s string, sl []string) bool {
	for i := range sl {
		if sl[i] == s {
			return true
		}
	}
	return false
}
