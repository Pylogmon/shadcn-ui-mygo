// Package fonts provides the 26 font families offered by the fixed shadcn
// Create source, plus optional fonts used by individual component examples.
// Calling a family function links only that family's assets; registration is
// explicit. Package fonts/all is the separate full Create catalogue.
package fonts

import (
	"errors"
	"fmt"
	"sync"

	"github.com/egoist/mygo/ui"
)

//go:generate go run ./internal/generate

type Kind string

const (
	Sans  Kind = "sans"
	Mono  Kind = "mono"
	Serif Kind = "serif"
)

// Info describes a family without requiring its binary assets. Weights lists
// the source-supported 100-step weights; native matching handles intermediate
// requests. Instrument Serif supplies only its original regular 400 face.
type Info struct {
	ID, Name, Family string
	Kind             Kind
	Weights          []int
	Italic           bool
}

// License is the unchanged upstream license or accompanying copyright notice.
// Include these notices when distributing an app that embeds the font.
type License struct {
	Path, Text string
}

// Font is a family descriptor. It holds immutable asset/registration functions,
// never UI state. Register it before constructing text that uses Family().
type Font struct {
	info     Info
	register func() error
	licenses func() []License
}

func (f Font) Info() Info {
	i := f.info
	i.Weights = append([]int(nil), i.Weights...)
	return i
}

// Family is the name accepted by Element.Font, Theme.Font and theme.FontRoles.
func (f Font) Family() string { return f.info.Family }

// Register adds the family's normal and, when available, italic faces through
// MyGo's public API. Calls are safe to repeat and share one result.
// No font is registered merely by importing this package.
func (f Font) Register() error {
	if f.register == nil {
		return errors.New("fonts: no font selected")
	}
	return f.register()
}

func (f Font) Licenses() []License {
	if f.licenses == nil {
		return nil
	}
	return f.licenses()
}

type registration struct {
	once sync.Once
	err  error
}

func (r *registration) add(family string, assets ...[]byte) error {
	r.once.Do(func() {
		for _, data := range assets {
			if err := ui.RegisterFont(data, family); err != nil {
				r.err = fmt.Errorf("fonts: register %s: %w", family, err)
				return
			}
		}
	})
	return r.err
}
