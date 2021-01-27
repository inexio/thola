package utility

// IfThenElse is a wrapper for the if condition.
func IfThenElse(condition bool, t interface{}, e interface{}) interface{} {
	if condition {
		return t
	}
	return e
}

// IfThenElseInt is a wrapper for the if condition.
func IfThenElseInt(condition bool, t int, e int) int {
	if condition {
		return t
	}
	return e
}

// IfThenElseString is a wrapper for the if condition.
func IfThenElseString(condition bool, t string, e string) string {
	if condition {
		return t
	}
	return e
}

func SliceUniqueString(a []string) []string {
	l := len(a)
	seen := make(map[string]struct{}, l)
	k := 0

	for i := 0; i < l; i++ {
		v := a[i]
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		a[k] = v
		k++
	}

	return a[0:k]
}

func SliceUniqueInt(a []int) []int {
	l := len(a)
	seen := make(map[int]struct{}, l)
	k := 0

	for i := 0; i < l; i++ {
		v := a[i]
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		a[k] = v
		k++
	}

	return a[0:k]
}
