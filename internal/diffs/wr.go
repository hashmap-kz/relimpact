package diffs

import (
	"fmt"
	"io"
)

func xFprintf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
