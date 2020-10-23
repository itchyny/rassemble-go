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
	return alternate(ra.rs...).String(), nil
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
		return addLiteral(r1, r2.Rune)
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
						addLiteral(literal([]rune{r2.Rune[0]}), r2.Rune),
						chars(append(r1.Rune[:j], r1.Rune[j+2:]...)),
					)
				}
			}
		}
	}
	return nil
}

func addLiteral(r *syntax.Regexp, runes []rune) *syntax.Regexp {
	switch r.Op {
	case syntax.OpLiteral:
		if i := compareRunes(r.Rune, runes); i > 0 {
			if i == len(r.Rune) && i == len(runes) {
				return r
			} else if i == len(r.Rune) {
				return concat(literal(r.Rune), quest(literal(runes[i:])))
			} else if i == len(runes) {
				return concat(literal(runes), quest(literal(r.Rune[i:])))
			} else if i+1 == len(r.Rune) && i+1 == len(runes) {
				r1, r2 := r.Rune[i], runes[i]
				if r1 > r2 {
					r1, r2 = r2, r1
				}
				return concat(literal(runes[:i]), chars([]rune{r1, r1, r2, r2}))
			} else {
				return concat(
					literal(runes[:i]),
					alternate(literal(r.Rune[i:]), literal(runes[i:])),
				)
			}
		} else if len(r.Rune) == 1 && len(runes) == 1 {
			r1, r2 := r.Rune[0], runes[0]
			if r1 > r2 {
				r1, r2 = r2, r1
			}
			return chars([]rune{r1, r1, r2, r2})
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
				if i == len(r0.Rune) && i == len(runes) {
					if len(r.Sub) == 2 {
						switch r.Sub[1].Op {
						case syntax.OpQuest, syntax.OpStar:
							return r
						case syntax.OpPlus:
							return concat(literal(runes), star(r.Sub[1].Sub[0]))
						default:
							return concat(literal(runes), quest(r.Sub[1]))
						}
					}
					return concat(literal(runes), quest(concat(r.Sub[1:]...)))
				} else if i == len(r0.Rune) {
					if len(r.Sub) == 2 {
						switch r.Sub[1].Op {
						case syntax.OpAlternate:
							for j, rr := range r.Sub[1].Sub {
								if s := addLiteral(rr, runes[i:]); s != nil {
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
							if s := addLiteral(r.Sub[1].Sub[0], runes[i:]); s != nil {
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
				} else if i == len(runes) {
					return concat(
						literal(runes),
						quest(concat(append([]*syntax.Regexp{literal(r0.Rune[i:])}, r.Sub[1:]...)...)),
					)
				} else {
					return concat(
						literal(runes[:i]),
						alternate(
							concat(append([]*syntax.Regexp{literal(r0.Rune[i:])}, r.Sub[1:]...)...),
							literal(runes[i:]),
						),
					)
				}
			}
		}
	case syntax.OpAlternate:
		for i, rr := range r.Sub {
			if s := addLiteral(rr, runes); s != nil {
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

func concat(sub ...*syntax.Regexp) *syntax.Regexp {
	return &syntax.Regexp{Op: syntax.OpConcat, Sub: sub}
}

func alternate(sub ...*syntax.Regexp) *syntax.Regexp {
	return &syntax.Regexp{Op: syntax.OpAlternate, Sub: sub}
}

func literal(runes []rune) *syntax.Regexp {
	return &syntax.Regexp{Op: syntax.OpLiteral, Rune: runes}
}

func quest(re *syntax.Regexp) *syntax.Regexp {
	return &syntax.Regexp{Op: syntax.OpQuest, Sub: []*syntax.Regexp{re}}
}

func star(re *syntax.Regexp) *syntax.Regexp {
	return &syntax.Regexp{Op: syntax.OpStar, Sub: []*syntax.Regexp{re}}
}

func chars(runes []rune) *syntax.Regexp {
	if len(runes) == 2 && runes[0] == runes[1] {
		return literal([]rune{runes[0]})
	}
	return &syntax.Regexp{Op: syntax.OpCharClass, Rune: runes}
}
