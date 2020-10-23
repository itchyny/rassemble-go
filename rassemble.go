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
	ra.rs = append(ra.rs, r)
	return nil
}

func (ra *rassemble) String() string {
	return (&syntax.Regexp{Op: syntax.OpAlternate, Sub: ra.rs}).String()
}
