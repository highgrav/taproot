package validation

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	YYYYMMDD         string = "2006-01-30"
	YYYYMMDDTZHHMMSS string = "2006-01-30T12:30:01"
)

type Validator struct {
	Errors map[string]string
}

func (v *Validator) IsValid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key, val string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = val
	}
}

func (v *Validator) Check(ok bool, key, val string) {
	if !ok {
		v.AddError(key, val)
	}
}

func (v *Validator) IsDate(fmt, val string, key, errmsg string) *Validator {
	_, err := time.Parse(fmt, val)
	v.Check(err == nil, key, errmsg)
	return v
}

func (v *Validator) IsNotBlank(val string, key, errmsg string) *Validator {
	v.Check(strings.TrimSpace(val) != "", key, errmsg)
	return v
}

func (v *Validator) IsWebURL(val string, key, errmsg string) *Validator {
	u, err := url.Parse(val)
	v.Check(err == nil, key, errmsg)
	v.Check(u.Scheme != "", key, errmsg)
	v.Check(u.Host != "", key, errmsg)
	v.Check(u.Scheme == "http" || u.Scheme == "https", key, errmsg)
	return v
}

func (v *Validator) IsInt(val string, key, errmsg string) *Validator {
	_, err := strconv.ParseInt(val, 10, 64)
	v.Check(err == nil, key, errmsg)
	return v
}

func (v *Validator) SliceLengthLTE(val []any, maxLen int, key, errmsg string) *Validator {
	v.Check(len(val) <= maxLen, key, errmsg)
	return v
}

func (v *Validator) SliceLengthGTE(val []any, minLen int, key, errmsg string) *Validator {
	v.Check(len(val) >= minLen, key, errmsg)
	return v
}

func (v *Validator) StrLengthLTE(val string, maxLen int, key, errmsg string) *Validator {
	v.Check(len(val) <= maxLen, key, errmsg)
	return v
}

func (v *Validator) StrLengthGTE(val string, minLen int, key, errmsg string) *Validator {
	v.Check(len(val) >= minLen, key, errmsg)
	return v
}

func (v *Validator) CheckFor(ok bool, key, errmsg string) *Validator {
	v.Check(ok, key, errmsg)
	return v
}

func (v *Validator) Matches(val string, rx *regexp.Regexp, key, errmsg string) *Validator {
	v.Check(rx.MatchString(val), key, errmsg)
	return v
}

func AreUnique[T comparable](vals []T) bool {
	uv := make(map[T]bool)
	for _, v := range vals {
		uv[v] = true
	}
	return len(vals) == len(uv)
}

func New() *Validator {
	return &Validator{
		Errors: make(map[string]string),
	}
}
