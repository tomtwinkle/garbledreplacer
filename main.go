package garbledreplacer

import (
	"errors"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/transform"
)

var ErrInvalidUTF8 = errors.New("invalid UTF-8 character")

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

	for len(_src) > 0 {
		r, size := utf8.DecodeRune(_src)
		if r < utf8.RuneSelf {
			size = 1
		} else if size == 1 {
			// All valid runes of size 1 (those below utf8.RuneSelf) were
			// handled above. We have invalid UTF-8, or we haven't seen the
			// full character yet.
			if !atEOF && !utf8.FullRune(_src) {
				err = transform.ErrShortSrc
				break
			}
			//　If the last string cannot be converted to rune, it is not replaced.
			if atEOF && !utf8.FullRune(_src) {
				break
			}
		}
		buf := _src[:size]
		if _, encErr := t.enc.Bytes(buf); encErr != nil {
			// Replace strings that cannot be converted
			buf = []byte(string(t.replaceRune))
		}
		if nDst+len(buf) > len(dst) {
			// over destination buffer
			err = transform.ErrShortDst
			break
		}
		dstN := copy(dst[nDst:], buf)
		if dstN <= 0 {
			break
		}
		nSrc += size
		nDst += dstN
		_src = _src[size:]
	}
	return
}
