package language

import (
	"slices"
	"strings"
	"unsafe"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"golang.org/x/text/transform"

	"github.com/lufia/godoc2man/internal/ascii"
)

var jaTokenizer *tokenizer.Tokenizer

func init() {
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}
	jaTokenizer = t
}

type Japanese struct {
}

func (j *Japanese) Reset() {
}

func (j *Japanese) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for {
		p, err := takeParagraph(src, atEOF)
		if err != nil {
			return nDst, nSrc, err
		}
		if len(p) == 0 {
			return nDst, nSrc, nil
		}
		s := j.breakString(string(p))
		if len(s) > len(dst) {
			return nDst, nSrc, transform.ErrShortDst
		}
		b := unsafe.Slice(unsafe.StringData(s), len(s))
		copy(dst, b)
		nSrc += len(p)
		nDst += len(b)
		src = src[len(p):]
		dst = dst[len(b):]
	}
}

func (j *Japanese) breakString(s string) string {
	var buf strings.Builder
	tokens := jaTokenizer.Tokenize(s)
	for _, token := range tokens {
		buf.WriteString(token.Surface)
		if j.canBreakAfter(token.Features()) {
			buf.WriteByte(ascii.UnitSeparator)
		}
	}
	return buf.String()
}

func (*Japanese) canBreakAfter(features []string) bool {
	switch {
	default:
		return false
	case slices.Contains(features, "読点"):
		return true
	case slices.Contains(features, "句点"):
		return true
	case slices.Contains(features, "助詞") && !slices.Contains(features, "連体化"):
		return true
	}
}
