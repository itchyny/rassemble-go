package rassemble

import "regexp/syntax"

// Join patterns and build a regexp pattern.
func Join(patterns []string) (string, error) {
	var rs []*syntax.Regexp
	var err error
	for _, pattern := range patterns {
		if rs, err = add(rs, pattern); err != nil {
			return "", err
		}
	}
	return mergeSuffix(alternate(rs...)).String(), nil
}

func add(rs []*syntax.Regexp, pattern string) ([]*syntax.Regexp, error) {
	r2, err := syntax.Parse(pattern, syntax.PerlX)
	if err != nil {
		return nil, err
	}
	for i, r1 := range rs {
		if r := merge0(r1, r2); r != nil {
			return insert(rs, r, i), nil
		}
	}
	for i, r1 := range rs {
		if r := merge1(r1, r2); r != nil {
			return insert(rs, r, i), nil
		}
	}
	if r2.Op == syntax.OpAlternate {
		return append(rs, r2.Sub...), nil
	}
	return append(rs, r2), nil
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
			// x+|(?:) => x*
			r1.Op = syntax.OpStar
			return r1
		case syntax.OpEmptyMatch, syntax.OpQuest, syntax.OpStar:
			// (?:)|(?:) => (?:)
			// x?|(?:) => x?
			// x*|(?:) => x*
			return r1
		}
	case syntax.OpPlus:
		if r1.Op == syntax.OpEmptyMatch {
			// (?:)|x+ => x*
			r2.Op = syntax.OpStar
			return r2
		}
	case syntax.OpQuest, syntax.OpStar:
		if r1.Op == syntax.OpEmptyMatch {
			// (?:)|x? => x?
			// (?:)|x* => x*
			return r2
		}
	case syntax.OpLiteral:
		return mergeLiteral(r1, r2.Rune)
	case syntax.OpCharClass:
		if r1.Op == syntax.OpLiteral && len(r1.Rune) == 1 {
			// a|[bc] => [a-c]
			return chars(addCharClass(r2.Rune, r1.Rune[0]))
		}
	case syntax.OpConcat:
		return mergeConcat(r1, r2.Sub)
	}
	if r1.Op == syntax.OpConcat && r2.Equal(r1.Sub[0]) {
		// x*y*z*|x* => x*(?:y*z*)?
		return concat(r2, quest(concat(r1.Sub[1:]...)))
	}
	return nil
}

func merge1(r1, r2 *syntax.Regexp) *syntax.Regexp {
	switch r2.Op {
	case syntax.OpEmptyMatch:
		switch r1.Op {
		case syntax.OpLiteral:
			// abc|(?:) => (?:abc)?
			return quest(r1)
		case syntax.OpCharClass:
			// [a-c]|(?:) => [a-c]?
			return quest(r1)
		}
	case syntax.OpCharClass:
		if r1.Op == syntax.OpEmptyMatch {
			// (?:)|[a-c] => [a-c]?
			return quest(r2)
		}
	case syntax.OpLiteral:
		switch r1.Op {
		case syntax.OpEmptyMatch:
			// (?:)|abc => (?:abc)?
			return quest(r2)
		case syntax.OpCharClass:
			for j := 0; j < len(r1.Rune); j += 2 {
				if r1.Rune[j] == r1.Rune[j+1] && r1.Rune[j] == r2.Rune[0] {
					// [acd]|ab => ab?|[cd]
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
				// abc|abc => abc
				return r
			}
			// abcd|abef => ab(?:cd|ef)
			return concat(
				literal(runes[:i]),
				alternate(literal(r.Rune[i:]), literal(runes[i:])),
			)
		} else if len(r.Rune) == 1 && len(runes) == 1 {
			// a|b => [ab]
			return alternate(r, literal(runes))
		}
	case syntax.OpCharClass:
		if len(runes) == 1 {
			// [a-c]|d => [a-d]
			return chars(addCharClass(r.Rune, runes[0]))
		}
	case syntax.OpConcat:
		r0 := r.Sub[0]
		switch r0.Op {
		case syntax.OpLiteral:
			if i := compareRunes(r0.Rune, runes); i > 0 {
				if i == len(r0.Rune) {
					if i == len(runes) {
						// abcx*y*|abc => abc(?:x*y*)?
						return concat(literal(runes), quest(concat(r.Sub[1:]...)))
					} else if len(r.Sub) == 2 {
						switch r.Sub[1].Op {
						case syntax.OpAlternate:
							for j, rr := range r.Sub[1].Sub {
								if s := mergeLiteral(rr, runes[i:]); s != nil {
									// abc(?:de|fg)|abcd => abc(de?|fg)
									r.Sub[1].Sub[j] = s
									return r
								}
							}
						case syntax.OpQuest:
							if s := mergeLiteral(r.Sub[1].Sub[0], runes[i:]); s != nil {
								// abc(?:d)?|abcde => abc(?:de?)?
								r.Sub[1].Sub[0] = s
								return r
							}
						}
					}
					// abcx*y*|abcde => abc(?:x*y*|de)
					return concat(
						literal(r0.Rune),
						alternate(concat(r.Sub[1:]...), literal(runes[i:])),
					)
				}
				// abcdx*y*|abef => ab(?:cdx*y*|ef)
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
				// (?:ab|cd)|cdef => ab|cd(?:ef)?
				r.Sub[i] = s
				return r
			}
		}
		return alternate(append(r.Sub, literal(runes))...)
	case syntax.OpQuest:
		if len(runes) == 1 && r.Sub[0].Op == syntax.OpCharClass {
			// [a-c]?|d => [a-d]?
			return quest(chars(addCharClass(r.Sub[0].Rune, runes[0])))
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
				for i >= 2 && rs[i-1]+1 == rs[i] {
					rs = append(rs[:i-1], rs[i+1:]...)
					i -= 2
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
		// x*|x*y*z* => x*(?:y*z*)?
		return concat(r, quest(concat(rs[1:]...)))
	}
	if r.Op == syntax.OpConcat {
		var i int
		for ; i < len(r.Sub) && i < len(rs); i++ {
			if !r.Sub[i].Equal(rs[i]) {
				if i > 0 {
					// x*y*z*w*|x*y*u*v* => x*y*(?:z*w*|u*v*)
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
				// x*y*|x*y* => x*y*
				return r
			}
			// x*y*|x*y*z*w* => x*y*(?:z*w*)?
			return concat(append(r.Sub, quest(concat(rs[i:]...)))...)
		} else if i == len(rs) {
			// x*y*z*w*|x*y* => x*y*(?:z*w*)?
			return concat(append(rs, quest(concat(r.Sub[i:]...)))...)
		}
		if r.Sub[0].Op == syntax.OpLiteral && rs[0].Op == syntax.OpLiteral {
			rs1, rs2 := r.Sub[0].Rune, rs[0].Rune
			if i := compareRunes(rs1, rs2); i > 0 {
				r.Sub[0], rs[0] = literal(rs1[i:]), literal(rs2[i:])
				// abcdx*|abefy* => ab(?:cdx*|efy*)
				return concat(
					literal(rs1[:i]),
					alternate(concat(r.Sub...), concat(rs...)),
				)
			}
		}
	}
	return nil
}

func mergeSuffix(r *syntax.Regexp) *syntax.Regexp {
	switch r.Op {
	case syntax.OpAlternate:
		return alternate(mergeSuffices(r.Sub)...)
	case syntax.OpQuest:
		if r.Sub[0].Op == syntax.OpAlternate {
			for j, rr := range r.Sub[0].Sub {
				if rr.Op == syntax.OpLiteral {
					s := quest(literal(rr.Rune))
					for _, rr := range r.Sub[0].Sub {
						if rr.Op == syntax.OpConcat && s.Equal(rr.Sub[len(rr.Sub)-1]) {
							r.Sub[0].Sub[j] = s
							// (?:ab?|b)? => (?:ab?|b?) => a?b?
							return mergeSuffix(r.Sub[0])
						}
					}
				}
			}
		}
		fallthrough
	case syntax.OpConcat, syntax.OpStar, syntax.OpPlus, syntax.OpRepeat:
		for i, rr := range r.Sub {
			r.Sub[i] = mergeSuffix(rr)
		}
		if r.Op == syntax.OpConcat {
			for _, rr := range r.Sub {
				if rr.Op == syntax.OpConcat {
					sub := make([]*syntax.Regexp, 0, len(r.Sub)+1)
					for _, rr := range r.Sub {
						if rr.Op == syntax.OpConcat {
							sub = append(sub, rr.Sub...)
						} else {
							sub = append(sub, rr)
						}
					}
					return concat(sub...)
				}
			}
		}
		return r
	default:
		return r
	}
}

func mergeSuffices(rs []*syntax.Regexp) []*syntax.Regexp {
	for i := 0; i < len(rs); i++ {
		rs[i] = mergeSuffix(rs[i])
	}
	for i := 0; i < len(rs)-1; i++ {
		r1 := rs[i]
		for j := i + 1; j < len(rs); j++ {
			r2 := rs[j]
			switch r1.Op {
			case syntax.OpLiteral:
				switch r2.Op {
				case syntax.OpLiteral:
					if k := compareRunesReverse(r1.Rune, r2.Rune); k > 0 {
						// abcd|cdcd => (?:ab|cd)cd
						r1 = concat(
							alternate(
								literal(r1.Rune[:len(r1.Rune)-k]),
								literal(r2.Rune[:len(r2.Rune)-k]),
							),
							literal(r1.Rune[len(r1.Rune)-k:]),
						)
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
							// x*cd|abcd => (?:x*|ab)cd
							r1 = concat(
								alternate(
									concat(append(r1.Sub[:len(r1.Sub)-1], literal(rs1[:len(rs1)-k]))...),
									literal(r2.Rune[:len(r2.Rune)-k]),
								),
								literal(r2.Rune[len(r2.Rune)-k:]),
							)
							rs = append(rs[:j], rs[j+1:]...)
							j--
						}
					}
				case syntax.OpConcat:
					var merged bool
					for k, l := len(r1.Sub)-1, len(r2.Sub)-1; k >= 0 && l >= 0; k, l = k-1, l-1 {
						if !r1.Sub[k].Equal(r2.Sub[l]) {
							if k < len(r1.Sub)-1 {
								// abx*y*z*|cdw*y*z* => (?:abx*|cdw*)y*z*
								r1 = concat(
									append(
										[]*syntax.Regexp{alternate(concat(r1.Sub[:k+1]...), concat(r2.Sub[:l+1]...))},
										r1.Sub[k+1:]...,
									)...,
								)
								rs = append(rs[:j], rs[j+1:]...)
								j--
								merged = true
							}
							break
						}
					}
					if !merged && r1.Sub[len(r1.Sub)-1].Op == syntax.OpLiteral &&
						r2.Sub[len(r2.Sub)-1].Op == syntax.OpLiteral {
						rs1, rs2 := r1.Sub[len(r1.Sub)-1].Rune, r2.Sub[len(r2.Sub)-1].Rune
						if k := compareRunesReverse(rs1, rs2); k > 0 {
							// x*abcd|y*cdcd => (?:x*ab|y*cd)cd
							r1 = concat(
								alternate(
									concat(append(r1.Sub[:len(r1.Sub)-1], literal(rs1[:len(rs1)-k]))...),
									concat(append(r2.Sub[:len(r2.Sub)-1], literal(rs2[:len(rs2)-k]))...),
								),
								literal(rs1[len(rs1)-k:]),
							)
							rs = append(rs[:j], rs[j+1:]...)
							j--
						}
					}
				default:
					if r2.Equal(r1.Sub[len(r1.Sub)-1]) {
						// x*y*z*|z* => (?:x*y*)?z*
						r1 = concat(quest(concat(r1.Sub[:len(r1.Sub)-1]...)), r2)
						rs = append(rs[:j], rs[j+1:]...)
						j--
					}
				}
			default:
				if r2.Op == syntax.OpConcat {
					if r1.Equal(r2.Sub[len(r2.Sub)-1]) {
						// z*|x*y*z* => (?:x*y*)?z*
						r1 = concat(quest(concat(r2.Sub[:len(r2.Sub)-1]...)), r1)
						rs = append(rs[:j], rs[j+1:]...)
						j--
					}
				}
			}
		}
		if r1 != rs[i] {
			rs[i] = mergeSuffix(r1)
		}
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
	switch len(sub) {
	case 1:
		return sub[0]
	case 2:
		if sub[1].Op == syntax.OpLiteral && len(sub[1].Rune) == 0 {
			// x*(?:) => x*
			return sub[0]
		}
	default:
		if sub[len(sub)-1].Op == syntax.OpLiteral && len(sub[len(sub)-1].Rune) == 0 {
			// x*y*(?:) => x*y*
			return &syntax.Regexp{Op: syntax.OpConcat, Sub: sub[:len(sub)-1]}
		}
	}
	return &syntax.Regexp{Op: syntax.OpConcat, Sub: sub}
}

func alternate(sub ...*syntax.Regexp) *syntax.Regexp {
	switch len(sub) {
	case 1:
		return sub[0]
	case 2:
		if sub[0].Op == syntax.OpLiteral && len(sub[0].Rune) == 1 &&
			sub[1].Op == syntax.OpLiteral && len(sub[1].Rune) == 1 {
			r1, r2 := sub[0].Rune[0], sub[1].Rune[0]
			if r1 > r2 {
				r1, r2 = r2, r1
			}
			return chars([]rune{r1, r1, r2, r2})
		} else if sub[0].Op == syntax.OpLiteral && len(sub[0].Rune) == 0 {
			// (?:)|x*y* => (?:x*y*)?
			return quest(sub[1])
		} else if sub[1].Op == syntax.OpLiteral && len(sub[1].Rune) == 0 {
			// x*y*|(?:) => (?:x*y*)?
			return quest(sub[0])
		} else {
			switch sub[0].Op {
			case syntax.OpAlternate:
				// (?:x*|y*)|z* => x*|y*|z*
				return &syntax.Regexp{Op: syntax.OpAlternate, Sub: append(sub[0].Sub, sub[1])}
			case syntax.OpQuest:
				// x?|y* => (?:x|y*)?
				return quest(alternate(sub[0].Sub[0], sub[1]))
			case syntax.OpLiteral:
				if len(sub[0].Rune) == 1 && sub[1].Op == syntax.OpCharClass {
					// d|[a-c] => [a-d]
					return chars(addCharClass(sub[1].Rune, sub[0].Rune[0]))
				}
			case syntax.OpCharClass:
				if sub[1].Op == syntax.OpLiteral && len(sub[1].Rune) == 1 {
					// [a-c]|d => [a-d]
					return chars(addCharClass(sub[0].Rune, sub[1].Rune[0]))
				}
			}
		}
	}
	return &syntax.Regexp{Op: syntax.OpAlternate, Sub: sub}
}

func literal(runes []rune) *syntax.Regexp {
	return &syntax.Regexp{Op: syntax.OpLiteral, Rune: runes}
}

func quest(r *syntax.Regexp) *syntax.Regexp {
	switch r.Op {
	case syntax.OpQuest, syntax.OpStar:
		// (?:x?)? => x?
		// (?:x*)? => x*
		return r
	case syntax.OpPlus:
		// (?:x+)? => x*
		return &syntax.Regexp{Op: syntax.OpStar, Sub: r.Sub}
	default:
		return &syntax.Regexp{Op: syntax.OpQuest, Sub: []*syntax.Regexp{r}}
	}
}

func chars(runes []rune) *syntax.Regexp {
	if len(runes) == 2 && runes[0] == runes[1] {
		// [a-a] => a
		return literal([]rune{runes[0]})
	}
	return &syntax.Regexp{Op: syntax.OpCharClass, Rune: runes}
}
