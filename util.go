package mirango

func containsString(vals []string, a string) bool {
	for _, v := range vals {
		if a == v {
			return true
		}
	}
	return false
}

func stringsUnion(a []string, b []string) []string {
	for _, bb := range b {
		exists := false
		for _, aa := range a {
			if aa == bb {
				exists = true
				break
			}
		}
		if !exists {
			a = append(a, bb)
		}
	}
	return a
}

func middlewareUnion(a []Middleware, b []Middleware) []Middleware {
	for _, bb := range b {
		exists := false
		for _, aa := range a {
			if aa == bb {
				exists = true
				break
			}
		}
		if !exists {
			a = append(a, bb)
		}
	}
	return a
}

func middlewareAppend(a []Middleware, bb Middleware) []Middleware {
	exists := false
	for _, aa := range a {
		if aa == bb {
			exists = true
			break
		}
	}
	if !exists {
		a = append(a, bb)
	}
	return a
}
