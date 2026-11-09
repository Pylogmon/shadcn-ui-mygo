package conversation

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

//go:embed assets/message-scroller-arrow-down.svg
var messageScrollerArrowDownSVG []byte

var messageScrollerArrowDown = ui.MustParseSVG(messageScrollerArrowDownSVG)

func messageScrollerArrowDownIcon() *ui.SVG { return messageScrollerArrowDown }
