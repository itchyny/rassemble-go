package rassemble

import "regexp/syntax"

// Join patterns and build a regexp pattern.
func Join(patterns []string) (string, error) {
	ra := &rassemble{}
	for _, pattern := range patterns {
		if err := ra.add(pattern); err != nil {
			return "", err
		}
	}
	return mergeSuffix(alternate(ra.rs...)).String(), nil
}

type rassemble struct {
	rs []*syntax.Regexp
}

func (ra *rassemble) add(pattern string) error {
	r2, err := syntax.Parse(pattern, syntax.PerlX)
	if err != nil {
		return err
	}
	var added bool
	for i, r1 := range ra.rs {
		if r := merge0(r1, r2); r != nil {
			ra.rs = insert(ra.rs, r, i)
			added = true
			break
		}
	}
	if !added {
		for i, r1 := range ra.rs {
			if r := merge1(r1, r2); r != nil {
				ra.rs = insert(ra.rs, r, i)
				added = true
				break
			}
		}
	}
	if !added {
		if r2.Op == syntax.OpAlternate {
			ra.rs = append(ra.rs, r2.Sub...)
		} else {
			ra.rs = append(ra.rs, r2)
		}
	}
	return nil
}

func insert(rs []*syntax.Regexp, r *syntax.Regexp, i int) []*syntax.Regexp {
	if r.Op == syntax.OpAlternate {
		rs = append(rs, r.Sub[1:]...)
		copy(rs[i+len(r.Sub):], rs[i+1:])
		copy(rs[i:], r.Sub)
	} else {
		rs[i] = r
	}
	return rs
}

func merge0(r1, r2 *syntax.Regexp) *syntax.Regexp {
	switch r2.Op {
	case syntax.OpEmptyMatch:
		switch r1.Op {
		case syntax.OpPlus:
			r1.Op = syntax.OpStar
			fallthrough
		case syntax.OpEmptyMatch, syntax.OpQuest, syntax.OpStar:
			return r1
		}
	case syntax.OpLiteral:
		return mergeLiteral(r1, r2.Rune)
	case syntax.OpConcat:
		return mergeConcat(r1, r2.Sub)
	default:
		if r1.Op == syntax.OpConcat && r2.Equal(r1.Sub[0]) {
			return concat(r2, quest(concat(r1.Sub[1:]...)))
		}
	}
	return nil
}

func merge1(r1, r2 *syntax.Regexp) *syntax.Regexp {
	switch r2.Op {
	case syntax.OpEmptyMatch:
		switch r1.Op {
		case syntax.OpLiteral:
			return quest(literal(r1.Rune))
		case syntax.OpCharClass:
			return quest(r1)
		}
	case syntax.OpLiteral:
		switch r1.Op {
		case syntax.OpCharClass:
			for j := 0; j < len(r1.Rune); j += 2 {
				if r1.Rune[j] == r1.Rune[j+1] && r1.Rune[j] == r2.Rune[0] {
					return alternate(
						mergeLiteral(literal([]rune{r2.Rune[0]}), r2.Rune),
						chars(append(r1.Rune[:j], r1.Rune[j+2:]...)),
					)
				}
			}
		}
	}
	return nil
}

func mergeLiteral(r *syntax.Regexp, runes []rune) *syntax.Regexp {
	switch r.Op {
	case syntax.OpLiteral:
		if i := compareRunes(r.Rune, runes); i > 0 {
			if i == len(r.Rune) && i == len(runes) {
				return r
			}
			return concat(literal(runes[:i]), literals(r.Rune[i:], runes[i:]))
		} else if len(r.Rune) == 1 && len(runes) == 1 {
			return literals(r.Rune, runes)
		}
	case syntax.OpCharClass:
		if len(runes) == 1 {
			r.Rune = addCharClass(r.Rune, runes[0])
			return r
		}
	case syntax.OpConcat:
		r0 := r.Sub[0]
		switch r0.Op {
		case syntax.OpLiteral:
			if i := compareRunes(r0.Rune, runes); i > 0 {
				if i == len(r0.Rune) {
					if i == len(runes) {
						return concat(literal(runes), quest(concat(r.Sub[1:]...)))
					} else if len(r.Sub) == 2 {
						switch r.Sub[1].Op {
						case syntax.OpAlternate:
							for j, rr := range r.Sub[1].Sub {
								if s := mergeLiteral(rr, runes[i:]); s != nil {
									r.Sub[1].Sub[j] = s
									return r
								}
							}
							return concat(
								literal(r0.Rune),
								alternate(append(r.Sub[1].Sub, literal(runes[i:]))...),
							)
						case syntax.OpCharClass:
							if i+1 == len(runes) {
								r.Sub[1].Rune = addCharClass(r.Sub[1].Rune, runes[i])
								return r
							}
						case syntax.OpQuest:
							if s := mergeLiteral(r.Sub[1].Sub[0], runes[i:]); s != nil {
								r.Sub[1].Sub[0] = s
								return r
							}
							return concat(
								literal(r0.Rune),
								quest(alternate(r.Sub[1].Sub[0], literal(runes[i:]))),
							)
						}
					}
					return concat(
						literal(r0.Rune),
						alternate(concat(r.Sub[1:]...), literal(runes[i:])),
					)
				}
				return concat(
					literal(runes[:i]),
					alternate(
						concat(append([]*syntax.Regexp{literal(r0.Rune[i:])}, r.Sub[1:]...)...),
						literal(runes[i:]),
					),
				)
			}
		}
	case syntax.OpAlternate:
		for i, rr := range r.Sub {
			if s := mergeLiteral(rr, runes); s != nil {
				r.Sub[i] = s
				return r
			}
		}
		return alternate(append(r.Sub, literal(runes))...)
	case syntax.OpQuest:
		if len(runes) == 1 && r.Sub[0].Op == syntax.OpCharClass {
			r.Sub[0].Rune = addCharClass(r.Sub[0].Rune, runes[0])
			return r
		}
	}
	return nil
}

func addCharClass(rs []rune, r rune) []rune {
	for i := 0; i < len(rs); i += 2 {
		if r < rs[i] {
			if r+1 == rs[i] {
				rs[i] = r
				if i+2 < len(rs) && rs[i+1]+1 == rs[i+2] {
					rs = append(rs[:i+1], rs[i+3:]...)
				}
				if i >= 2 && rs[i-1]+1 == r {
					rs = append(rs[:i-1], rs[i+1:]...)
				}
				return rs
			}
			rs = append(append(rs, 0, 0))
			copy(rs[i+2:], rs[i:])
			rs[i] = r
			rs[i+1] = r
			if i -= 4; i >= 0 && rs[i] == rs[i+1] &&
				rs[i]+1 == rs[i+2] && rs[i+2] == rs[i+3] && rs[i+2]+1 == r {
				rs = append(rs[:i+1], rs[i+5:]...)
			}
			return rs
		} else if r <= rs[i+1] {
			return rs
		} else if rs[i+1]+1 == r && rs[i] < rs[i+1] {
			rs[i+1] = r
			for i+2 < len(rs) && rs[i+2] == rs[i+1]+1 {
				rs = append(rs[:i+1], rs[i+3:]...)
			}
			return rs
		}
	}
	rs = append(append(rs, r), r)
	if i := len(rs) - 6; i >= 0 && rs[i] == rs[i+1] &&
		rs[i]+1 == rs[i+2] && rs[i+2] == rs[i+3] && rs[i+2]+1 == r {
		rs = append(rs[:i+1], r)
	}
	return rs
}

func mergeConcat(r *syntax.Regexp, rs []*syntax.Regexp) *syntax.Regexp {
	if r.Equal(rs[0]) {
		return concat(r, quest(concat(rs[1:]...)))
	}
	switch r.Op {
	case syntax.OpConcat:
		var i int
		for ; i < len(r.Sub) && i < len(rs); i++ {
			if !r.Sub[i].Equal(rs[i]) {
				if i > 0 {
					return concat(
						append(
							append([]*syntax.Regexp{}, rs[:i]...),
							alternate(concat(r.Sub[i:]...), concat(rs[i:]...)),
						)...,
					)
				}
				break
			}
		}
		if i == len(r.Sub) {
			if i == len(rs) {
				return r
			}
			return concat(append(r.Sub, quest(concat(rs[i:]...)))...)
		} else if i == len(rs) {
			return concat(append(rs, quest(concat(r.Sub[i:]...)))...)
		}
	}
	return nil
}

func mergeSuffix(r *syntax.Regexp) *syntax.Regexp {
	switch r.Op {
	case syntax.OpAlternate:
		return alternate(mergeSuffices(r.Sub)...)
	case syntax.OpConcat, syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		for i, rr := range r.Sub {
			r.Sub[i] = mergeSuffix(rr)
		}
		return r
	default:
		return r
	}
}

func mergeSuffices(rs []*syntax.Regexp) []*syntax.Regexp {
	for i := 0; i < len(rs)-1; i++ {
		r1 := rs[i]
		for j := i + 1; j < len(rs); j++ {
			r2 := rs[j]
			switch r1.Op {
			case syntax.OpLiteral:
				switch r2.Op {
				case syntax.OpLiteral:
					if k := compareRunesReverse(r1.Rune, r2.Rune); k > 0 {
						rs[i] = concat(
							literals(r1.Rune[:len(r1.Rune)-k], r2.Rune[:len(r2.Rune)-k]),
							literal(r1.Rune[len(r1.Rune)-k:]),
						)
						r1 = rs[i]
						rs = append(rs[:j], rs[j+1:]...)
						j--
					}
				}
			case syntax.OpConcat:
				switch r2.Op {
				case syntax.OpLiteral:
					if r1.Sub[len(r1.Sub)-1].Op == syntax.OpLiteral {
						rs1 := r1.Sub[len(r1.Sub)-1].Rune
						if k := compareRunesReverse(rs1, r2.Rune); k > 0 {
							if k == len(rs1) {
								rs[i] = concat(
									alternate(
										concat(r1.Sub[:len(r1.Sub)-1]...),
										literal(r2.Rune[:len(r2.Rune)-k]),
									),
									literal(r2.Rune[len(r2.Rune)-k:]),
								)
							} else {
								rs[i] = concat(
									alternate(
										concat(append(r1.Sub[:len(r1.Sub)-1], literal(rs1[:len(rs1)-k]))...),
										literal(r2.Rune[:len(r2.Rune)-k]),
									),
									literal(r2.Rune[len(r2.Rune)-k:]),
								)
							}
							r1 = rs[i]
							rs = append(rs[:j], rs[j+1:]...)
							j--
						}
					}
				case syntax.OpConcat:
					for k, l := len(r1.Sub)-1, len(r2.Sub)-1; k >= 0 && l >= 0; k, l = k-1, l-1 {
						if !r1.Sub[k].Equal(r2.Sub[l]) {
							if k < len(r1.Sub)-1 {
								rs[i] = concat(
									append(
										[]*syntax.Regexp{alternate(concat(r1.Sub[:k+1]...), concat(r2.Sub[:l+1]...))},
										r1.Sub[k+1:]...,
									)...,
								)
								r1 = rs[i]
								rs = append(rs[:j], rs[j+1:]...)
								j--
							}
							break
						}
					}
				default:
					if r2.Equal(r1.Sub[len(r1.Sub)-1]) {
						rs[i] = concat(quest(concat(r1.Sub[:len(r1.Sub)-1]...)), r2)
						r1 = rs[i]
						rs = append(rs[:j], rs[j+1:]...)
						j--
					}
				}
			default:
				switch r2.Op {
				case syntax.OpConcat:
					if r1.Equal(r2.Sub[len(r2.Sub)-1]) {
						rs[i] = concat(quest(concat(r2.Sub[:len(r2.Sub)-1]...)), r1)
						r1 = rs[i]
						rs = append(rs[:j], rs[j+1:]...)
						j--
					}
				}
			}
		}
		rs[i] = mergeSuffix(r1)
	}
	return rs
}

func compareRunes(xs, ys []rune) int {
	var i int
	for _, y := range ys {
		if i == len(xs) || xs[i] != y {
			break
		}
		i++
	}
	return i
}

func compareRunesReverse(xs, ys []rune) int {
	var i int
	for i < len(ys) {
		if i == len(xs) || xs[len(xs)-1-i] != ys[len(ys)-1-i] {
			break
		}
		i++
	}
	return i
}

func concat(sub ...*syntax.Regexp) *syntax.Regexp {
	if len(sub) == 1 {
		return sub[0]
	}
	return &syntax.Regexp{Op: syntax.OpConcat, Sub: sub}
}

func alternate(sub ...*syntax.Regexp) *syntax.Regexp {
	switch len(sub) {
	case 1:
		return sub[0]
	case 2:
		if sub[0].Op == syntax.OpLiteral && len(sub[0].Rune) == 0 {
			return quest(sub[1])
		} else if sub[1].Op == syntax.OpLiteral && len(sub[1].Rune) == 0 {
			return quest(sub[0])
		}
	}
	return &syntax.Regexp{Op: syntax.OpAlternate, Sub: sub}
}

func literal(runes []rune) *syntax.Regexp {
	return &syntax.Regexp{Op: syntax.OpLiteral, Rune: runes}
}

func literals(rs1, rs2 []rune) *syntax.Regexp {
	if len(rs1) == 1 && len(rs2) == 1 {
		r1, r2 := rs1[0], rs2[0]
		if r1 > r2 {
			r1, r2 = r2, r1
		}
		return chars([]rune{r1, r1, r2, r2})
	}
	return alternate(literal(rs1), literal(rs2))
}

func quest(r *syntax.Regexp) *syntax.Regexp {
	switch r.Op {
	case syntax.OpQuest, syntax.OpStar:
		return r
	case syntax.OpPlus:
		return &syntax.Regexp{Op: syntax.OpStar, Sub: r.Sub}
	default:
		return &syntax.Regexp{Op: syntax.OpQuest, Sub: []*syntax.Regexp{r}}
	}
}

func chars(runes []rune) *syntax.Regexp {
	if len(runes) == 2 && runes[0] == runes[1] {
		return literal([]rune{runes[0]})
	}
	return &syntax.Regexp{Op: syntax.OpCharClass, Rune: runes}
}
