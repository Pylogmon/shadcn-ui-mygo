package conversation

import (
	"embed"
	"strings"

	"github.com/egoist/mygo/ui"
)

// Original Lucide SVGs; provenance and hashes are in assets/attachment-manifest.json.
// The shared ISC/Feather license is bundled from internal/indicator.
//
//go:embed assets/attachment-*.svg
var attachmentAssets embed.FS

var attachmentIcons = func() map[string]*ui.SVG {
	entries, err := attachmentAssets.ReadDir("assets")
	if err != nil {
		panic(err)
	}
	icons := make(map[string]*ui.SVG, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		source, err := attachmentAssets.ReadFile("assets/" + name)
		if err != nil {
			panic(err)
		}
		icons[strings.TrimSuffix(strings.TrimPrefix(name, "attachment-"), ".svg")] = ui.MustParseSVG(source)
	}
	return icons
}()

func attachmentDefaultIcon(kind string) *ui.SVG {
	if icon := attachmentIcons[kind]; icon != nil {
		return icon
	}
	return attachmentIcons["file-text"]
}
