package desktop

import "github.com/egoist/mygo/ui"

type contextHostKey struct{}
type contextHost struct{ host *Host }

// BindHost associates a Host with this window's Context. Call it at the start
// of the window view, alongside BeginFrame/EndFrame. Nil clears the binding.
// A separate native child Context has its own binding; there is no global Host.
func BindHost(c *ui.Context, host *Host) {
	if c == nil || !c.Valid() {
		return
	}
	ui.Local(c.Root(), contextHostKey{}, func() contextHost { return contextHost{} }).host = host
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
