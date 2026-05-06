package diffs

import (
	"fmt"
	"io"
)

// hfprintf writes formatted HTML to w, discarding write errors (the caller
// is building an in-memory string builder and errors are impossible there).
func hfprintf(w io.Writer, format string, a ...any) {
	_, _ = fmt.Fprintf(w, format, a...)
}
