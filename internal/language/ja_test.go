package language

import (
	"errors"
	"io"
	"strings"
	"testing"

	"golang.org/x/text/transform"
)

func TestJapaneseTransform_noError(t *testing.T) {
	t.Parallel()

	const N = 9
	tests := map[string]string{
		"":         "",         // 0 bytes
		"1":        "1",        // 1 byte
		"すもも":      "すもも",      // 9 bytes, just N
		"1\n2\n":   "1\n2\n",   // 4 bytes, having newlines
		"1\n\n2\n": "1\n\n2\n", // 4 bytes, having a block
	}
	for s, want := range tests {
		var j Japanese
		buf := make([]byte, N)
		nDst, nSrc, err := j.Transform(buf, []byte(s), true)
		if err != nil {
			t.Fatalf("Transform(%s): %v", s, err)
		}
		if nDst != len(want) {
			t.Errorf("nDst = %d (%q); want %d (%q)", nDst, buf, len(want), want)
		}
		if nSrc != len(s) {
			t.Errorf("nSrc = %d; want %d (%q)", nDst, len(s), s)
		}
	}
}

func TestJapaneseTransform_shortDst(t *testing.T) {
	t.Parallel()

	const N = 9
	tests := map[string]struct {
		nDst, nSrc int
	}{
		"すもも1": {0, 0}, // 10byte, valid as utf8
		"はい。":  {0, 0}, // 9byte, valid as utf8, + `\:`
	}
	for s, tt := range tests {
		t.Run(s, func(t *testing.T) {
			var j Japanese
			buf := make([]byte, N)
			r := newTransformResult(j.Transform(buf, []byte(s), true))
			want := newTransformResult(tt.nDst, tt.nSrc, transform.ErrShortDst)
			testTransformResult(t, r, want)
		})
	}
}

func TestJapaneseTransform_shortSrcEOF(t *testing.T) {
	t.Parallel()

	const N = 9
	tests := map[string]struct {
		nDst, nSrc int
	}{
		"すもも\xe3": {0, 0}, // 10byte, invalid as utf8
	}
	for s, tt := range tests {
		t.Run(s, func(t *testing.T) {
			var j Japanese
			buf := make([]byte, N)
			r := newTransformResult(j.Transform(buf, []byte(s), true))
			want := newTransformResult(tt.nDst, tt.nSrc, transform.ErrShortSrc)
			testTransformResult(t, r, want)
		})
	}
}

func TestJapaneseTransform_shortSrcNotEOF(t *testing.T) {
	t.Parallel()

	const N = 9
	tests := map[string]struct {
		nDst, nSrc int
	}{
		"すもも": {0, 0}, // 10byte, invalid as utf8
	}
	for s, tt := range tests {
		t.Run(s, func(t *testing.T) {
			var j Japanese
			buf := make([]byte, N)
			r := newTransformResult(j.Transform(buf, []byte(s), false))
			want := newTransformResult(tt.nDst, tt.nSrc, transform.ErrShortSrc)
			testTransformResult(t, r, want)
		})
	}
}

type transformResult struct {
	nDst, nSrc int
	err        error
}

func newTransformResult(nDst, nSrc int, err error) *transformResult {
	return &transformResult{nDst, nSrc, err}
}

func testTransformResult(t *testing.T, r, want *transformResult) {
	t.Helper()
	if !errors.Is(r.err, want.err) {
		t.Errorf("err = %v; want %v", r.err, want.err)
	}
	if r.nDst != want.nDst {
		t.Errorf("nDst = %d; want %d", r.nDst, want.nDst)
	}
	if r.nSrc != want.nSrc {
		t.Errorf("nSrc = %d; want %d", r.nDst, want.nSrc)
	}
}

func TestJapaneseTransform_tokenize(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"すもももももももものうち": `すももも\:ももも\:もものうち`,
	}

	for s, want := range tests {
		f := strings.NewReader(s)
		r := NewReader(f, "ja")
		if v := readString(t, r); v != want {
			t.Errorf("%q -> %q; want %q", s, v, want)
		}
	}
}

func maxRepeat(s, t string, nbytes int) (s1, t1 string) {
	n := len(s)
	if n == 0 {
		return s, t
	}
	r := (nbytes / n) + 1
	return strings.Repeat(s, r), strings.Repeat(t, r)
}

func readString(t *testing.T, r io.Reader) string {
	t.Helper()
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal("failed to read:", err)
	}
	return string(b)
}
