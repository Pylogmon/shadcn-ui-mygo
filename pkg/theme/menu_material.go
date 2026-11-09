package theme

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
	"sync"
)

// MenuBackdrop reports the actual rendering capability, not the requested look.
type MenuBackdrop uint8

const (
	MenuOpaque MenuBackdrop = iota
	MenuSceneBackdrop
	MenuSystemBackdrop
)

func (m MenuSnapshot) Backdrop(c *ui.Context) MenuBackdrop {
	p := c.Preferences()
	if !m.Translucent() || p.ReduceTransparency || p.HighContrast {
		return MenuOpaque
	}
	if m.nativeWindow {
		if c.Vibrancy() {
			return MenuSystemBackdrop
		}
		return MenuOpaque
	}
	return MenuSceneBackdrop
}
func (m MenuSnapshot) Surface(c *ui.Context, element ui.Element) ui.Element {
	element.TextColor(m.Tokens.PopoverForeground).Border(1, m.Tokens.Foreground.Alpha(.1))
	switch m.Backdrop(c) {
	case MenuSceneBackdrop:
		element.Material(menuMaterial{tint: m.Tokens.Popover.Alpha(.7)})
	case MenuSystemBackdrop:
		element.Background(m.Tokens.Popover.Alpha(.7))
	default:
		element.Background(m.Tokens.Popover)
	}
	return element
}

// The source's backdrop-blur-2xl is a 40DIP Gaussian; saturation is the
// CSS saturate(1.5) matrix, applied to the backdrop before the 70% tint.
// A Material draws beneath the children, so text and icons remain sharp.
const menuShaderSource = `package main
var Tint vec4
func Fragment(dst vec4, pos vec2) vec4 {
 b:=backdropSample(pos)
 rgb:=b.rgb/max(b.a,0.00001)
 l:=dot(rgb,vec3(0.213,0.715,0.072))
 rgb=clamp(vec3(l)+(rgb-vec3(l))*1.5,vec3(0),vec3(1))
 return vec4(Tint.rgb*Tint.a+rgb*b.a*(1-Tint.a),Tint.a+b.a*(1-Tint.a))
}`

var menuShader struct {
	sync.Once
	value *shader.Shader
}

func menuBackdropShader() *shader.Shader {
	menuShader.Do(func() {
		var err error
		menuShader.value, err = shader.Compile(menuShaderSource)
		if err != nil {
			panic(err)
		}
	})
	return menuShader.value
}

type menuMaterial struct{ tint ui.Color }

func (m menuMaterial) PaintMaterial(p *ui.Painter, box ui.Rect, radii [4]float32) {
	menuBackdropShader().Draw(p, box, shader.Options{BackdropBlur: 40, Radius: radii[0], Uniforms: map[string]any{"Tint": m.tint}})
}
