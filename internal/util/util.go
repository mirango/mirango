package util

func ContainsString(vals []string, a string) bool {
	for _, v := range vals {
		if a == v {
			return true
		}
	}
	return false
}
