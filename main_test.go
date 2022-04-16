package garbledreplacer_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/tomtwinkle/garbledreplacer"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func TestNewTransformer(t *testing.T) {
	tests := map[string]struct {
		encoding encoding.Encoding
		in       string
		replace  rune
		want     string
	}{
		"UTF-8->ShiftJIS:no garbled text": {
			encoding: japanese.ShiftJIS,
			in:       strings.Repeat("一二三四五六七八九十拾壱", 1000),
			replace:  '?',
			want:     strings.Repeat("一二三四五六七八九十拾壱", 1000),
		},
		"UTF-8->ShiftJIS:with garbled text": {
			encoding: japanese.ShiftJIS,
			in:       strings.Repeat("一二三四五六七八九十拾壱🍺", 1000),
			replace:  '?',
			want:     strings.Repeat("一二三四五六七八九十拾壱?", 1000),
		},
		"UTF-8->ShiftJIS:with garbled text:other replaceRune": {
			encoding: japanese.ShiftJIS,
			in:       strings.Repeat("一二三四🍣五六七八九🍺十拾壱", 1000),
			replace:  '@',
			want:     strings.Repeat("一二三四@五六七八九@十拾壱", 1000),
		},
	}

	for n, v := range tests {
		name := n
		tt := v

		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			w := transform.NewWriter(&buf, garbledreplacer.NewTransformer(tt.encoding, tt.replace))
			if _, err := w.Write([]byte(tt.in)); err != nil {
				panic(err)
			}
			if err := w.Close(); err != nil {
				panic(err)
			}

			var actual bytes.Buffer
			aw := transform.NewWriter(&actual, tt.encoding.NewDecoder())
			if _, err := aw.Write(buf.Bytes()); err != nil {
				panic(err)
			}
			if err := aw.Close(); err != nil {
				panic(err)
			}

			if len([]rune(tt.want)) != len([]rune(actual.String())) {
				t.Error("string length does not match")
			}
			if tt.want != actual.String() {
				t.Error("string does not match")
			}
		})
	}
}
