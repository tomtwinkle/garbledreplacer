package garbledreplacer

import (
	"bytes"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

func NewTransformer(enc encoding.Encoding, replaceRune rune) transform.Transformer {
	e := enc.NewEncoder()
	return transform.Chain(&replacer{
		replaceRune: replaceRune,
		enc:         e,
	}, e)
}

type replacer struct {
	transform.NopResetter

	enc         *encoding.Encoder
	replaceRune rune
}

var _ transform.Transformer = (*replacer)(nil)

func (t *replacer) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	_src := src

	if len(_src) == 0 && atEOF {
		return
	}

	var idx int
	if len(_src) < len(dst) {
		idx = len(_src)
	} else {
		idx = len(dst)
	}
	for _, r := range bytes.Runes(_src[:idx]) {
		if r == utf8.RuneError {
			continue
		}
		buf := []byte(string(r))
		nSrc += len(buf)
		if _, err := t.enc.Bytes(buf); err != nil {
			buf = []byte(string(t.replaceRune))
		}
		nDst += copy(dst[nDst:], buf)
	}

	if nDst < idx {
		err = transform.ErrShortDst
	}
	return
}
