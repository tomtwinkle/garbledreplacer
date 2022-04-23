package garbledreplacer_test

import (
	"bytes"
	"github.com/tomtwinkle/garbledreplacer"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"strings"
	"testing"
	"unicode/utf8"
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
			in:       strings.Repeat("一二三四🍣五六七八九🍺十拾壱", 3000),
			replace:  '@',
			want:     strings.Repeat("一二三四@五六七八九@十拾壱", 3000),
		},
		"UTF-8->ShiftJIS:RuneError only": {
			encoding: japanese.ShiftJIS,
			in:       string(utf8.RuneError),
			replace:  '?',
			want:     "?",
		},
		"UTF-8->EUCJP:with garbled text": {
			encoding: japanese.EUCJP,
			in:       strings.Repeat("一二三四🍣五六七八九🍺十拾壱", 3000),
			replace:  '?',
			want:     strings.Repeat("一二三四?五六七八九?十拾壱", 3000),
		},
		"UTF-8->Big5:with garbled text": {
			encoding: traditionalchinese.Big5,
			in:       strings.Repeat("咖呸咕咀呻🍣呷咄咒咆呼咐🍺呱呶和咚呢", 3000),
			replace:  '?',
			want:     strings.Repeat("咖呸咕咀呻?呷咄咒咆呼咐?呱呶和咚呢", 3000),
		},
	}

	for n, v := range tests {
		name := n
		tt := v

		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			w := transform.NewWriter(&buf, garbledreplacer.NewTransformer(tt.encoding, tt.replace))
			if _, err := w.Write([]byte(tt.in)); err != nil {
				t.Error(err)
			}
			if err := w.Close(); err != nil {
				t.Error(err)
			}

			var actual bytes.Buffer
			aw := transform.NewWriter(&actual, tt.encoding.NewDecoder())
			if _, err := aw.Write(buf.Bytes()); err != nil {
				t.Error(err)
			}
			if err := aw.Close(); err != nil {
				t.Error(err)
			}

			if len([]rune(tt.want)) != len([]rune(actual.String())) {
				t.Errorf("string length does not match %d=%d", len([]rune(tt.want)), len([]rune(actual.String())))
			}
			if tt.want != actual.String() {
				t.Errorf("string does not match\n%s", actual.String())
			}
		})
	}
}

func FuzzTransformer(f *testing.F) {
	f.Skip()
	seeds := [][]byte{
		bytes.Repeat([]byte("一二三四五六七八九十拾壱🍺"), 1000),
		bytes.Repeat([]byte("一二三四🍣五六七八九🍺十拾壱"), 3000),
		bytes.Repeat([]byte("一二三四🍣五六七八九🍺十拾壱"), 3000),
		bytes.Repeat([]byte("咖呸咕咀呻🍣呷咄咒咆呼咐🍺呱呶和咚呢"), 3000),
	}

	for _, b := range seeds {
		f.Add(b)
	}

	f.Fuzz(func(t *testing.T, p []byte) {
		tr := garbledreplacer.NewTransformer(japanese.ShiftJIS, '?')
		for len(p) > 0 {
			_, n, err := transform.Bytes(tr, p)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			p = p[n:]
		}
	})
}
