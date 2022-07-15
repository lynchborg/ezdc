package dchar

func gmap[T any, U any](t []T, f func(t T) U) (res []U) {
	for _, item := range t {
		out := f(item)
		res = append(res, out)
	}
	return
}

func filter[T any](t []T, f func(t T) bool) (res []T) {
	for _, item := range t {
		if f(item) {
			res = append(res, item)
		}
	}
	return
}

func find[T comparable](t []T, f T) bool {
	for _, item := range t {
		if item == f {
			return true
		}
	}
	return false
}
