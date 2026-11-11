package gallery

import "github.com/egoist/mygo/ui"

// The source directory complements the existing interactive explorers. Its
// state belongs to this Gallery draft; menu/theme scopes remain outside it.
func (s *State) chartSourcePreview(c *ui.Context) {
	s.chartCatalogue.Initialize()
	s.chartCatalogue.Translate = tr
	s.chartCatalogue.View(c)
}
