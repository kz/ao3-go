package ao3

import (
	"github.com/microcosm-cc/bluemonday"
	"errors"
)

// ao3Policy is the sanitizer for AO3's Limited HTML.
//
// A current list of known HTML tags used on AO3 are: b, strong, i em, strike,
// s, del, u, ins, sub, sup, big, small, tt, pre, code, kbd, samp, var, address,
// cite and q. <p> and <br> are implicitly allowed.
//
// More info: https://archiveofourown.org/works/5191202/chapters/11961779
//
// The above tags known to not be supported on Android are: ins, pre, code, kbd,
// samp, var, address and q.
//
// The full list of tags supported on Android can be found here:
// https://android.googlesource.com/platform/frameworks/base/+/master/core/java/android/text/Html.java

type SanitizationPolicy int

const (
	// NonePolicy instructs the sanitizer not to perform any sanitization
	NonePolicy SanitizationPolicy = 0
	// AO3Policy instructs the sanitizer to only keep AO3's limited HTML tags
	AO3Policy SanitizationPolicy = 1
	// AO3AndroidPolicy instructs the sanitizer to keep the AO3Policy tags
	// which are supported by Android's TextView
	AO3AndroidPolicy SanitizationPolicy = 2
)

type Sanitizer struct {
	sanitizer *bluemonday.Policy
}

func NewSanitizer(strength SanitizationPolicy) (*Sanitizer, error) {
	var allowedTags []string
	if strength == NonePolicy {
		return &Sanitizer{sanitizer: nil}, nil
	} else if strength == AO3Policy {
		allowedTags = []string{"p", "br", "b", "strong", "i", "em", "strike", "s", "del", "u", "ins", "sub", "sup", "big", "small", "tt", "pre", "code", "kbd", "samp", "var", "address", "cite", "q"}
	} else if strength == AO3AndroidPolicy {
		allowedTags = []string{"p", "br", "b", "strong", "i", "em", "strike", "s", "del", "u", "sub", "sup", "big", "small", "tt", "cite"}
	} else {
		return nil, errors.New("invalid sanitizer strength")
	}

	p := bluemonday.NewPolicy()
	p.AllowStandardURLs()

	for _, tag := range allowedTags {
		p.AllowElements(tag)
	}

	p.AllowStandardAttributes()

	return &Sanitizer{sanitizer: p}, nil
}

func (sanitizer *Sanitizer) Sanitize(html string) string {
	if sanitizer.sanitizer == nil {
		return html
	} else {
		return sanitizer.sanitizer.Sanitize(html)
	}
}