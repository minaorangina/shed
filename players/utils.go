package players

func charsUnique(s string) bool {
	seen := map[string]bool{}
	for _, c := range s {
		if _, ok := seen[string(c)]; ok {
			return false
		}
		seen[string(c)] = true
	}
	return true
}

func charsInRange(chars string, lower, upper int) bool {
	for _, char := range chars {
		if int(char) < lower || int(char) > upper {
			return false
		}
	}
	return true
}
