package eval

import (
	"fmt"
	"regexp"

	"github.com/samber/lo"
)

func headed(vals []string) []string {
	return append([]string{"HEAD"}, vals...)
}

func filter(rx string, elems []string) (res []string, err error) {
	r, err := regexp.Compile(rx)
	if err != nil {
		return nil, fmt.Errorf("compile regexp: %w", err)
	}

	for _, e := range elems {
		if r.MatchString(e) {
			res = append(res, e)
		}
	}

	return res, nil
}

func next(elem string, elems []string) string {
	for idx, e := range elems {
		if e == elem {
			if idx+1 == len(elems) {
				return ""
			}

			return elems[idx+1]
		}
	}

	return ""
}

func previous(elem string, elems []string) string {
	for idx, e := range elems {
		if e == elem {
			if idx == 0 {
				return ""
			}

			return elems[idx-1]
		}
	}

	return ""
}

func strings(elems []interface{}) []string {
	out, _ := lo.FromAnySlice[string](elems)
	return out
}
