package language

import (
	"bytes"
	"errors"
	"io"
	"unicode/utf8"

	"golang.org/x/text/transform"
)

var transformers = map[string]transform.Transformer{
	"ja": &Japanese{},
}

// String returns the string transformed by the transformer corresponding to lang.
func String(lang, s string) (string, error) {
	t, ok := transformers[lang]
	if !ok {
		return s, nil
	}
	v, _, err := transform.String(t, s)
	return v, err
}

var blank = []byte{'\n'}

// takeParagraph returns any bytes up to blank line, including it, or to end of src.
func takeParagraph(src []byte, atEOF bool) ([]byte, error) {
	n := 0
	for {
		line, err := takeLine(src[n:], atEOF)
		if err != nil {
			return nil, err
		}
		n += len(line)
		if len(line) == 0 || bytes.Equal(line, blank) {
			return src[:n], nil
		}
	}
}

// takeLine returns any bytes up to '\n', including it, or to end of src.
func takeLine(src []byte, atEOF bool) ([]byte, error) {
	var (
		r = &utf8Reader{src, 0}
		n = 0
	)
	for {
		c, size, err := r.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				if !atEOF {
					return nil, transform.ErrShortSrc
				}
				return src[:n], nil
			}
			return nil, err
		}
		n += size
		if c == '\n' {
			return src[:n], nil
		}
	}
}

type utf8Reader struct {
	src []byte
	p   int
}

func (r *utf8Reader) ReadRune() (c rune, size int, err error) {
	rest := r.src[r.p:]
	if len(rest) == 0 {
		return 0, 0, io.EOF
	}
	c, n := utf8.DecodeRune(rest)
	if c == utf8.RuneError && len(rest) < utf8.UTFMax {
		return 0, 0, transform.ErrShortSrc
	}
	r.p += n
	return c, n, nil
}
