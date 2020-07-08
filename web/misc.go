package web

func stringIn(str string, candidates []string) bool {
	for _, c := range candidates {
		if c == str {
			return true
		}
	}

	return false
}

