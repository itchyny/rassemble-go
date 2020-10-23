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
	r, err := syntax.Parse(pattern, syntax.POSIX)
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
		var i int
		for _, c := range runes {
			if i == len(r.Rune) || r.Rune[i] != c {
				break
			}
			i++
		}
		if i > 0 {
			s := &syntax.Regexp{
				Op: syntax.OpConcat,
				Sub: []*syntax.Regexp{
					{
						Op:   syntax.OpLiteral,
						Rune: runes[:i],
					},
				},
			}
			if i == len(r.Rune) {
				if i != len(runes) {
					s.Sub = append(s.Sub, &syntax.Regexp{
						Op: syntax.OpQuest,
						Sub: []*syntax.Regexp{
							{
								Op:   syntax.OpLiteral,
								Rune: runes[i:],
							},
						},
					})
				}
			} else if i == len(runes) {
				s.Sub = append(s.Sub, &syntax.Regexp{
					Op: syntax.OpQuest,
					Sub: []*syntax.Regexp{
						{
							Op:   syntax.OpLiteral,
							Rune: r.Rune[i:],
						},
					},
				})
			} else {
				s.Sub = append(s.Sub, &syntax.Regexp{
					Op: syntax.OpAlternate,
					Sub: []*syntax.Regexp{
						{
							Op:   syntax.OpLiteral,
							Rune: r.Rune[i:],
						},
						{
							Op:   syntax.OpLiteral,
							Rune: runes[i:],
						},
					},
				})
			}
			return s
		}
	}
	return nil
}

func (ra *rassemble) String() string {
	return (&syntax.Regexp{Op: syntax.OpAlternate, Sub: ra.rs}).String()
}
