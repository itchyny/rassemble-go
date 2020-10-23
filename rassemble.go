package rassemble

import "regexp/syntax"

// Join patterns and build a regexp pattern.
func Join(patterns []string) (string, error) {
	ra := &rassemble{rs: make([]*syntax.Regexp, 0, len(patterns))}
	for _, pattern := range patterns {
		if err := ra.add(pattern); err != nil {
			return "", err
		}
	}
	return ra.String(), nil
}

type rassemble struct {
	rs []*syntax.Regexp
}

func (ra *rassemble) add(pattern string) error {
	r, err := syntax.Parse(pattern, syntax.PerlX)
	if err != nil {
		return err
	}
	var added bool
	switch r.Op {
	case syntax.OpLiteral:
		for i, rr := range ra.rs {
			if s := addLiteral(rr, r.Rune); s != nil {
				ra.rs[i], added = s, true
				break
			}
		}
	}
	if !added {
		ra.rs = append(ra.rs, r)
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
			} else {
				return concat(
					literal(runes[:i]),
					alternate(literal(r.Rune[i:]), literal(runes[i:])),
				)
			}
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
	}
	return nil
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

func (ra *rassemble) String() string {
	return alternate(ra.rs...).String()
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
