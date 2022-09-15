// Package rassemble provides a method to assemble regular expressions.
package rassemble

import (
	"regexp/syntax"
	"sort"
	"unicode"
)

// Join patterns to build a regexp pattern.
func Join(patterns []string) (string, error) {
	var sub []*syntax.Regexp
	for _, pattern := range patterns {
		r, err := syntax.Parse(pattern, syntax.PerlX|syntax.ClassNL)
		if err != nil {
			return "", err
		}
		sub = add(sub, breakLiterals(r))
	}
	return mergeSuffix(alternate(sub...)).String(), nil
}

func breakLiterals(r *syntax.Regexp) *syntax.Regexp {
	switch r.Op {
	case syntax.OpLiteral:
		if len(r.Rune) <= 1 {
			return r
		}
		sub := make([]*syntax.Regexp, len(r.Rune))
		for i := range r.Rune {
			sub[i] = &syntax.Regexp{Op: syntax.OpLiteral, Rune: r.Rune[i : i+1]}
		}
		return concat(sub...)
	default:
		for i, rr := range r.Sub {
			r.Sub[i] = breakLiterals(rr)
		}
		if r.Op == syntax.OpConcat {
			r = flattenConcat(r)
		}
		return r
	}
}

func add(sub []*syntax.Regexp, r2 *syntax.Regexp) []*syntax.Regexp {
	if r2.Op == syntax.OpAlternate {
		for _, r2 := range r2.Sub {
			sub = add(sub, r2)
		}
		return sub
	}
	for i, r1 := range sub {
		if r1.Equal(r2) {
			return sub
		}
		if r := mergePrefix(r1, r2); r != nil {
			sub[i] = r
			return sub
		}
	}
	return append(sub, r2)
}

func mergePrefix(r1, r2 *syntax.Regexp) *syntax.Regexp {
	if r1.Op > r2.Op {
		r1, r2 = r2, r1
	}
	switch r1.Op {
	case syntax.OpEmptyMatch:
		switch r2.Op {
		case syntax.OpLiteral, syntax.OpCharClass,
			syntax.OpStar, syntax.OpPlus, syntax.OpQuest:
			// (?:)|x+ => x*, etc.
			return quest(r2)
		}
	case syntax.OpLiteral:
		switch r2.Op {
		case syntax.OpCharClass:
			// a|[bc] => [a-c]
			return charClass(append(r2.Rune, r1.Rune[0], r1.Rune[0]))
		case syntax.OpQuest:
			if r2 := r2.Sub[0]; r2.Op == syntax.OpCharClass {
				// a|[bc]? => [a-c]?
				return quest(charClass(append(r2.Rune, r1.Rune[0], r1.Rune[0])))
			}
		}
	case syntax.OpCharClass:
		switch r2.Op {
		case syntax.OpCharClass:
			// [a-c]|[d-f] => [a-f]
			return charClass(append(r1.Rune, r2.Rune...))
		case syntax.OpQuest:
			switch r2 := r2.Sub[0]; r2.Op {
			case syntax.OpLiteral:
				// [ab]|c? => [a-c]?
				return quest(charClass(append(r1.Rune, r2.Rune[0], r2.Rune[0])))
			case syntax.OpCharClass:
				// [ab]|[cd]? => [a-d]?
				return quest(charClass(append(r1.Rune, r2.Rune...)))
			}
		}
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest:
		if r1.Sub[0].Equal(r2) {
			// x*|x => x*
			// x+|x => x+
			// x?|x => x?
			return r1
		}
		if r1.Op < r2.Op && r2.Op <= syntax.OpQuest && r1.Sub[0].Equal(r2.Sub[0]) {
			// x*|x+ => x*
			// x*|x? => x*
			// x+|x? => x*
			return &syntax.Regexp{Op: syntax.OpStar, Sub: r1.Sub}
		}
	case syntax.OpConcat:
		return mergePrefixConcat(r1, r2)
	}
	switch r2.Op {
	case syntax.OpConcat:
		return mergePrefixConcat(r2, r1)
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest:
		if r1.Equal(r2.Sub[0]) {
			// x|x* => x*
			// x|x? => x?
			// x|x+ => x+
			return r2
		}
	}
	return nil
}

func mergePrefixConcat(r1, r2 *syntax.Regexp) *syntax.Regexp {
	if r2.Op == syntax.OpConcat {
		var i int
		for ; i < len(r1.Sub) && i < len(r2.Sub); i++ {
			if !r1.Sub[i].Equal(r2.Sub[i]) {
				break
			}
		}
		if i > 0 {
			// x*y*z*w*|x*y*u*v* => x*y*(?:z*w*|u*v*)
			return concat(
				append(
					append(make([]*syntax.Regexp, 0, i+1), r1.Sub[:i]...),
					alternate(concat(r1.Sub[i:]...), concat(r2.Sub[i:]...)),
				)...,
			)
		}
	} else if r1.Sub[0].Equal(r2) {
		// x*y*z*|x* => x*(?:y*z*)?
		return concat(r2, quest(concat(r1.Sub[1:]...)))
	}
	return nil
}

func mergeSuffix(r *syntax.Regexp) *syntax.Regexp {
	for i, rr := range r.Sub {
		r.Sub[i] = mergeSuffix(rr)
	}
	switch r.Op {
	case syntax.OpAlternate:
		sub, k, rs := r.Sub, -1, r.Rune0[:0]
		for i := 0; i < len(sub); i++ {
			r1 := sub[i]
			for j := i + 1; j < len(sub); j++ {
				r2 := sub[j]
				if r := mergeSuffixConcat(r1, r2); r != nil {
					r1, j, sub = r, j-1, append(sub[:j], sub[j+1:]...)
				}
			}
			if r1 != sub[i] {
				sub[i] = mergeSuffix(r1)
				continue
			}
			// merge literals and character classes here
			// to prefer ax?|bx?|cx? over [abc]|ax|bx|cx
			switch r1.Op {
			case syntax.OpLiteral:
				rs = append(rs, r1.Rune[0], r1.Rune[0])
			case syntax.OpCharClass:
				rs = append(rs, r1.Rune...)
			default:
				continue
			}
			if k < 0 {
				k = i
			} else {
				i, sub = i-1, append(sub[:i], sub[i+1:]...)
			}
		}
		if k >= 0 && len(rs) > 2 {
			// (?:a|b|[c-e]) => [a-e]
			sub[k] = charClass(rs)
		}
		return alternate(sub...)
	case syntax.OpQuest:
		if r := r.Sub[0]; r.Op == syntax.OpAlternate {
			for i, rr := range r.Sub {
				if rr.Op == syntax.OpLiteral {
					for _, rs := range r.Sub {
						if rs.Op == syntax.OpConcat &&
							rs.Sub[len(rs.Sub)-1].Op == syntax.OpQuest &&
							rr.Equal(rs.Sub[len(rs.Sub)-1].Sub[0]) {
							// (?:ab?|b)? => (?:ab?|b?) => a?b?
							r.Sub[i] = quest(rr)
							return mergeSuffix(r)
						}
					}
				}
			}
		}
		return r
	case syntax.OpConcat:
		return flattenConcat(r)
	default:
		return r
	}
}

func mergeSuffixConcat(r1, r2 *syntax.Regexp) *syntax.Regexp {
	if r1.Op != syntax.OpConcat {
		if r2.Op != syntax.OpConcat {
			return nil
		}
		r1, r2 = r2, r1
	}
	if r2.Op == syntax.OpConcat {
		var i int
		for ; i < len(r1.Sub) && i < len(r2.Sub); i++ {
			if !r1.Sub[len(r1.Sub)-1-i].Equal(r2.Sub[len(r2.Sub)-1-i]) {
				break
			}
		}
		if i > 0 {
			// x*y*z*w*|u*v*z*w* => (?:x*y*|u*v*)z*w*
			return concat(
				append(
					[]*syntax.Regexp{
						alternate(
							concat(r1.Sub[:len(r1.Sub)-i]...),
							concat(r2.Sub[:len(r2.Sub)-i]...),
						),
					},
					r1.Sub[len(r1.Sub)-i:]...,
				)...,
			)
		}
	} else if r1.Sub[len(r1.Sub)-1].Equal(r2) {
		// x*y*z*|z* => (?:x*y*)?z*
		return concat(quest(concat(r1.Sub[:len(r1.Sub)-1]...)), r2)
	}
	return nil
}

func flattenConcat(r *syntax.Regexp) *syntax.Regexp {
	n := len(r.Sub)
	for _, rr := range r.Sub {
		if rr.Op == syntax.OpConcat {
			n += len(rr.Sub) - 1
		}
	}
	sub := make([]*syntax.Regexp, 0, n)
	for _, rr := range r.Sub {
		if rr.Op == syntax.OpConcat {
			sub = append(sub, rr.Sub...)
		} else {
			sub = append(sub, rr)
		}
	}
	return concat(sub...)
}

func concat(sub ...*syntax.Regexp) *syntax.Regexp {
	switch len(sub) {
	case 0:
		return &syntax.Regexp{Op: syntax.OpEmptyMatch}
	case 1:
		return sub[0]
	default:
		return &syntax.Regexp{Op: syntax.OpConcat, Sub: sub}
	}
}

func alternate(sub ...*syntax.Regexp) *syntax.Regexp {
	switch len(sub) {
	case 1:
		return sub[0]
	case 2:
		r1, r2 := sub[0], sub[1]
		if r := mergePrefix(r1, r2); r != nil {
			return r
		}
		if r2.Op == syntax.OpEmptyMatch {
			// x*y*|(?:) => (?:x*y*)?
			return quest(r1)
		}
		switch r1.Op {
		case syntax.OpEmptyMatch:
			// (?:)|x*y* => (?:x*y*)?
			return quest(r2)
		case syntax.OpAlternate:
			// (?:x*|y*)|z* => x*|y*|z*
			return alternate(add(r1.Sub, r2)...)
		case syntax.OpQuest:
			// x?|y* => (?:x|y*)?
			return quest(alternate(r1.Sub[0], r2))
		}
		fallthrough
	default:
		return &syntax.Regexp{Op: syntax.OpAlternate, Sub: sub}
	}
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
	case syntax.OpAlternate:
		for i, rr := range r.Sub {
			switch rr.Op {
			case syntax.OpQuest, syntax.OpStar:
				// (?:x|y?|z)? => x|y?|z
				// (?:x|y*|z)? => x|y*|z
				return r
			case syntax.OpPlus:
				// (?:x|y+|z)? => x|y*|z
				r.Sub[i].Op = syntax.OpStar
				return r
			}
		}
		fallthrough
	default:
		return &syntax.Regexp{Op: syntax.OpQuest, Sub: []*syntax.Regexp{r}}
	}
}

type charClassSlice []rune

func (rs charClassSlice) Len() int {
	return len(rs) / 2
}
func (rs charClassSlice) Less(i, j int) bool {
	return rs[i*2] < rs[j*2]
}
func (rs charClassSlice) Swap(i, j int) {
	i, j = i*2, j*2
	rs[i], rs[i+1], rs[j], rs[j+1] = rs[j], rs[j+1], rs[i], rs[i+1]
}

func charClass(rs []rune) *syntax.Regexp {
	sort.Sort(charClassSlice(rs))
	var i int
	for j := 2; j < len(rs); j += 2 {
		switch {
		case rs[i+1] >= rs[j]:
			if rs[i+1] < rs[j+1] {
				// [a-dc-e] => [a-e]
				rs[i+1] = rs[j+1]
			}
		case rs[i+1]+1 == rs[j]:
			switch {
			case i > 0 && rs[i-1]+1 == rs[i]:
				// [abc-e] => [a-e]
				i -= 2
				fallthrough
			case rs[i] < rs[i+1] || rs[j] < rs[j+1]:
				// [a-de], [ab-e] => [a-e]
				rs[i+1] = rs[j+1]
				continue
			}
			// [ab] =/> [a-b]
			fallthrough
		default:
			if i += 2; i != j {
				rs[i], rs[i+1] = rs[j], rs[j+1]
			}
		}
	}
	rs = rs[:i+2]
	if len(rs) == 2 && rs[0] == 0 && rs[1] == unicode.MaxRune {
		// [^a]|a => (?s:.)
		return &syntax.Regexp{Op: syntax.OpAnyChar}
	}
	return &syntax.Regexp{Op: syntax.OpCharClass, Rune: rs}
}
