// Package chartcatalog shares fixed shadcn source examples across Gallery and
// runnable native apps. Rendering uses only public kit/native APIs.
package chartcatalog

type Options struct {
	Selected, SelectedSeries *int
	Animate                  bool
	Range, Series, Month     string
}
