package fp

func Map[E any, T any](ls []E, fn func(e E) T) []T {
	res := make([]T, 0, len(ls))
	for _, te := range ls {
		res = append(res, fn(te))
	}

	return res
}

func Contains[E comparable](ls []E, e E) bool {
	for _, te := range ls {
		if te == e {
			return true
		}
	}

	return false
}

func ContainsAny[E any](ls []E, e E, fn func(e1, e2 E) bool) bool {
	for _, te := range ls {
		if fn(e, te) {
			return true
		}
	}

	return false
}

func Filter[E any](ls []E, fn func(e E) bool) []E {
	res := make([]E, 0)

	for _, te := range ls {
		if fn(te) {
			res = append(res, te)
		}
	}

	return res
}

func Reduce[E any](ls []E, fn func(e1, e2 E) E) (E, bool) {
	var res E
	switch len(ls) {
	case 0:
		return res, false
	case 1:
		res = ls[0]
	default:
		res = ls[0]
		for i := 1; i < len(ls); i++ {
			res = fn(res, ls[i])
		}
	}

	return res, true
}

func Foreach[E any](ls []E, fn func(e E)) {
	for _, te := range ls {
		fn(te)
	}
}

func Group[E any](ls []E, fn func(e1, e2 E) bool, aggr func(e1, e2 E) E) []E {
	grps := make([][]E, 0)
	res := make([]E, 0)

	for _, te := range ls {
		isG := false
		for i, g := range grps {
			if fn(te, g[0]) {
				grps[i] = append(grps[i], te)
				isG = true
				break
			}
		}

		if !isG {
			grps = append(grps, []E{te})
		}
	}

	for _, g := range grps {
		switch len(g) {
		case 1:
			res = append(res, g[0])
		default:
			var te E
			for i, j := 0, 1; j < len(g); {
				if i == 0 {
					te = aggr(g[i], g[j])
					i++
					j++
					continue
				}
				te = aggr(te, g[j])
				j++
			}
			res = append(res, te)
		}
	}

	return res
}

func Accumulate[E any, T any](ls []E, fn func(t T, e E) T, def func(e E) T) (T, bool) {
	var res T

	switch len(ls) {
	case 0:
		return res, false
	case 1:
		res = def(ls[0])
	default:
		res = def(ls[0])
		for j := 1; j < len(ls); j++ {
			res = fn(res, ls[j])
		}
	}

	return res, true
}
