package desktop

import (
	"strconv"

	"github.com/egoist/mygo/ui"
	controltooltip "github.com/pylogmon/shadcn-ui-mygo/internal/tooltip"
)

type contextHostKey struct{}
type contextHost struct{ host *Host }

// BindHost associates a Host with this window's Context. Call it at the start
// of the window view, alongside BeginFrame/EndFrame. Nil clears the binding.
// Compound-control tooltips use the same Host and Tooltip defaults. A separate
// native child Context has its own binding; there is no global Host.
func BindHost(c *ui.Context, host *Host) {
	if c == nil || !c.Valid() {
		return
	}
	ui.Local(c.Root(), contextHostKey{}, func() contextHost { return contextHost{} }).host = host
	if host == nil {
		controltooltip.Bind(c, nil)
		return
	}
	controltooltip.Bind(c, func(c *ui.Context, anchor ui.Element, text string) error {
		return Tooltip(c, host, "control-tooltip-"+strconv.FormatUint(anchor.ID(), 10), anchor, text, TooltipProps{})
	})
}

// HostOf returns this window's bound Host during a live build, or nil.
func HostOf(c *ui.Context) *Host {
	if c == nil || !c.Valid() {
		return nil
	}
	return ui.Local(c.Root(), contextHostKey{}, func() contextHost { return contextHost{} }).host
}

func resolveHost(c *ui.Context, explicit *Host) *Host {
	if explicit != nil {
		return explicit
	}
	return HostOf(c)
}
