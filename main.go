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
	if !utf8.Valid(_src) {
		// If not a string, do not process
		err = ErrInvalidUTF8
		return
	}

	for len(_src) > 0 {
		_, n := utf8.DecodeRune(_src)
		buf := _src[:n]
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
		nSrc += n
		nDst += dstN
		_src = _src[n:]
	}
	return
}
