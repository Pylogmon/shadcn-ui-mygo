package conversation

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

//go:embed assets/trace-chevron-right.svg
var traceChevronSource []byte

var traceChevronIcon = ui.MustParseSVG(traceChevronSource)
