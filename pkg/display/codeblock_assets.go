package display

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

// Exact Lucide 0.474.0 assets; provenance and hashes are beside the SVGs.
// The shared ISC/Feather notice is shipped from internal/indicator/LICENSE.lucide.
//
//go:embed assets/code-copy.svg
var codeCopySource []byte

//go:embed assets/code-check.svg
var codeCheckSource []byte

var codeCopyIcon = ui.MustParseSVG(codeCopySource)
var codeCopiedIcon = ui.MustParseSVG(codeCheckSource)
