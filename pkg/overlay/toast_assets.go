package overlay

import (
	_ "embed"

	"github.com/egoist/mygo/ui"
)

// Exact pinned Lucide status SVGs and Sonner's CloseIcon geometry.
// Provenance and licenses are recorded beside the assets and in testdata/sonner.
//
//go:embed assets/toast-success.svg
var toastSuccessSource []byte

//go:embed assets/toast-info.svg
var toastInfoSource []byte

//go:embed assets/toast-warning.svg
var toastWarningSource []byte

//go:embed assets/toast-error.svg
var toastErrorSource []byte

//go:embed assets/toast-close.svg
var toastCloseSource []byte

var (
	toastSuccessIcon = ui.MustParseSVG(toastSuccessSource)
	toastInfoIcon    = ui.MustParseSVG(toastInfoSource)
	toastWarningIcon = ui.MustParseSVG(toastWarningSource)
	toastErrorIcon   = ui.MustParseSVG(toastErrorSource)
	toastCloseIcon   = ui.MustParseSVG(toastCloseSource)
)
