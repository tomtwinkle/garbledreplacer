package garbledreplacer_test

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/tomtwinkle/garbledreplacer"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

func TestNewTransformer(t *testing.T) {
	tests := map[string]struct {
		encoding  encoding.Encoding
		in        string
		replace   rune
		want      string
		wantError error
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
		"UTF-8->ShiftJIS:Invalid UTF-8 character": {
			encoding: japanese.ShiftJIS,
			in:       "\xe4",
			replace:  '?',
			want:     "",
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
		"UTF-8->ShiftJIS:with garbled text:larger than 4096bytes": {
			encoding: japanese.ShiftJIS,
			in:       strings.Repeat("一二三四🍣五六七八九🍺十拾壱", 4000),
			replace:  '?',
			want:     strings.Repeat("一二三四?五六七八九?十拾壱", 4000),
		},
		"UTF-8->ShiftJIS:all garbled text:larger than 4096bytes": {
			encoding: japanese.ShiftJIS,
			in:       strings.Repeat("🍣🍣🍣🍺🍣🍣🍣", 4000),
			replace:  '?',
			want:     strings.Repeat("???????", 4000),
		},
	}

	assertFunc := func(t *testing.T, want string, actual bytes.Buffer, decoder *encoding.Decoder) {
		var assertBuf bytes.Buffer
		aw := transform.NewWriter(&assertBuf, decoder)
		if _, err := aw.Write(actual.Bytes()); err != nil {
			t.Error(err)
		}
		if err := aw.Close(); err != nil {
			t.Error(err)
		}

		if len([]rune(want)) != len([]rune(assertBuf.String())) {
			t.Errorf("string length does not match %d=%d", len([]rune(want)), len([]rune(assertBuf.String())))
		}
		if want != assertBuf.String() {
			t.Errorf("string does not match\n%s", assertBuf.String())
		}
	}

	for n, v := range tests {
		name := n
		tt := v

		t.Run("[transform.NewWriter]"+name, func(t *testing.T) {
			var buf bytes.Buffer
			w := transform.NewWriter(&buf, garbledreplacer.NewTransformer(tt.encoding, tt.replace))
			_, err := w.Write([]byte(tt.in))
			if tt.wantError != nil {
				if err == nil {
					t.Errorf("want error %v, got nil", tt.wantError)
				}
				if errors.Is(err, tt.wantError) {
					return
				}
				t.Error(err)
			}
			if err := w.Close(); err != nil {
				t.Error(err)
			}
			assertFunc(t, tt.want, buf, tt.encoding.NewDecoder())
		})
		t.Run("[transform.NewWriter with bufio.NewWriter]"+name, func(t *testing.T) {
			var buf bytes.Buffer
			w := bufio.NewWriter(transform.NewWriter(&buf, garbledreplacer.NewTransformer(tt.encoding, tt.replace)))
			_, err := w.WriteString(tt.in)
			if tt.wantError != nil {
				if err == nil {
					t.Errorf("want error %v, got nil", tt.wantError)
				}
				if errors.Is(err, tt.wantError) {
					return
				}
				t.Error(err)
			}
			if err := w.Flush(); err != nil {
				t.Error(err)
			}
			assertFunc(t, tt.want, buf, tt.encoding.NewDecoder())
		})
	}
}

// nolint: typecheck
func FuzzTransformer(f *testing.F) {
	f.Skip()
	seeds := [][]byte{
		bytes.Repeat([]byte("一二三四五六七八九十拾壱🍣🍺"), 1000),
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
			if !utf8.Valid(p) {
				t.Skip()
			}
			_, n, err := transform.Bytes(tr, p)
			if err != nil {
				t.Fatal("unexpected error:", err)
			}
			p = p[n:]
		}
	})
}
