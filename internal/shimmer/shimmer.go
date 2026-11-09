// Package shimmer paints the pinned shimmer utility through native text
// coverage. It never replaces MyGo's shaping, layout or accessibility tree.
package shimmer

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/shader"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/motion"
)

type Props struct {
	Active bool
	// Base is the source currentColor, including alpha. Set the desired color
	// here rather than overriding TextColor after this helper while it is active.
	Base ui.Color
	// Font sets known font attributes and measures the default 3ch + 40px band.
	// Omitted attributes inherit normally; unknown ancestor font overrides can
	// change ch width, so pass the matching font or an explicit Spread.
	Font ui.Font
	// MeasureFont overrides only the 3ch estimate, preserving inherited native styles.
	MeasureFont                  *ui.Font
	Highlight                    *ui.Color
	Duration                     time.Duration
	Angle, Spread, SpreadPercent *float32
	Once, Reverse                bool
}
type stateKey struct{}
type state struct {
	active   bool
	start    time.Time
	duration time.Duration
}

func Text(c *ui.Context, text string, p Props) ui.Element {
	root := ui.Text(c, text).TextColor(p.Base)
	font := p.Font
	if font.Family != "" {
		root.Font(font.Family)
	} else {
		font.Family = c.Theme().Font
	}
	if font.Size > 0 && finite(font.Size) {
		root.FontSize(font.Size)
	} else {
		font.Size = c.Theme().FontSize
		if font.Size <= 0 || !finite(font.Size) {
			font.Size = 14
		}
	}
	if font.Weight != 0 {
		root.FontWeight(font.Weight)
	} else {
		font.Weight = 400
	}
	if p.Font.LineHeight > 0 && finite(p.Font.LineHeight) {
		root.LineHeight(p.Font.LineHeight)
	}
	if p.Font.Italic {
		root.Italic()
	}
	if p.Font.Features != "" {
		root.FontFeatures(p.Font.Features)
	}
	if p.Font.Lang != "" {
		root.Lang(p.Font.Lang)
	}
	if p.MeasureFont != nil {
		measured := *p.MeasureFont
		if measured.Family != "" {
			font.Family = measured.Family
		}
		if measured.Size > 0 && finite(measured.Size) {
			font.Size = measured.Size
		}
		if measured.Weight != 0 {
			font.Weight = measured.Weight
		}
		if measured.LineHeight > 0 {
			font.LineHeight = measured.LineHeight
		}
		if measured.Features != "" {
			font.Features = measured.Features
		}
	}
	s := ui.Local(root, stateKey{}, func() state { return state{} })
	if !p.Active || motion.Reduced(c) {
		s.active = false
		return root
	}
	duration := p.Duration
	invalid := false
	if duration < 0 {
		duration = 0
		invalid = true
	}
	if duration == 0 {
		duration = 2 * time.Second
	}
	angle := float32(20)
	if p.Angle != nil {
		if finite(*p.Angle) {
			angle = *p.Angle
		} else {
			invalid = true
		}
	}
	spread := float32(0)
	fraction := float32(-1)
	if p.SpreadPercent != nil && finite(*p.SpreadPercent) && *p.SpreadPercent >= 0 {
		fraction = *p.SpreadPercent / 100
	} else {
		if p.SpreadPercent != nil {
			invalid = true
		}
		if p.Spread != nil && finite(*p.Spread) && *p.Spread >= 0 {
			spread = *p.Spread
		} else {
			if p.Spread != nil {
				invalid = true
			}
			zero := ui.ShapeText("0", font, 0)
			width, _ := zero.Size()
			spread = 3*width + 40
		}
	}
	root.Invalid(invalid)
	if !s.active || s.duration != duration {
		s.active = true
		s.start = c.Now()
		s.duration = duration
	}
	elapsed := c.Now().Sub(s.start)
	phase := float32(0)
	if p.Once {
		phase = min(1, float32(elapsed)/float32(duration))
		if phase < 1 {
			c.AnimationFrame()
		}
	} else {
		phase = float32(elapsed%duration) / float32(duration)
		c.AnimationFrame()
	}
	// Both the RTL selector and shimmer-reverse set animation-direction:reverse.
	if (c.Direction() == ui.RightToLeft || p.Reverse) && !(p.Once && elapsed >= duration) {
		phase = 1 - phase
	}
	base, middle, highlight := colors(p.Base, p.Highlight, c.Theme().Dark)
	coverage := p.Base
	coverage.A = 255
	root.TextColor(coverage).Effect(sweep.Effect(shader.Options{Uniforms: map[string]any{
		"Base": base, "Middle": middle, "Highlight": highlight, "Phase": phase, "Spread": spread, "Fraction": fraction, "Angle": angle * math.Pi / 180,
	}}))
	return root
}
func finite(v float32) bool { return !math.IsNaN(float64(v)) && !math.IsInf(float64(v), 0) }

// The colors stay in perceptual coordinates until the final per-pixel gamut
// mapping. In particular a dark highlight is never clipped to sRGB before the
// source's OKLCH midpoint and premultiplied OKLab gradient are interpolated.
func colors(base ui.Color, override *ui.Color, dark bool) (b, m, h [4]float32) {
	l, c, hue := lch(base)
	alpha := float64(base.A) / 255
	hl, hc, hh, ha := l, c, hue, alpha*.2
	if dark {
		hl = math.Min(1, math.Max(.8, l+.4))
		ha = math.Min(1, alpha+.4)
	}
	if override != nil {
		hl, hc, hh = lch(*override)
		ha = float64(override.A) / 255
	}
	b = lab(l, c, hue, alpha)
	h = lab(hl, hc, hh, ha)
	total := (alpha + ha) * .5
	if total == 0 {
		return b, [4]float32{}, h
	}
	weight := alpha / (alpha + ha)
	if c < .00001 {
		hue = hh
	}
	if hc < .00001 {
		hh = hue
	}
	delta := math.Mod(hue-hh+540, 360) - 180
	m = lab(hl+(l-hl)*weight, hc+(c-hc)*weight, hh+delta*.5, total)
	return
}
func lab(l, c, h, a float64) [4]float32 {
	r := h * math.Pi / 180
	return [4]float32{float32(l), float32(c * math.Cos(r)), float32(c * math.Sin(r)), float32(a)}
}
func lch(color ui.Color) (l, c, h float64) {
	linear := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= .04045 {
			return x / 12.92
		}
		return math.Pow((x+.055)/1.055, 2.4)
	}
	r, g, b := linear(color.R), linear(color.G), linear(color.B)
	x := math.Cbrt(.4122214708*r + .5363325363*g + .0514459929*b)
	y := math.Cbrt(.2119034982*r + .6806995451*g + .1073969566*b)
	z := math.Cbrt(.0883024619*r + .2817188376*g + .6299787005*b)
	l = .2104542553*x + .793617785*y - .0040720468*z
	a := 1.9779984951*x - 2.428592205*y + .4505937099*z
	v := .0259040371*x + .7827717662*y - .808675766*z
	c = math.Hypot(a, v)
	h = math.Atan2(v, a) * 180 / math.Pi
	if h < 0 {
		h += 360
	}
	return
}

var sweep = shader.MustCompile(`package main
var Base vec4
var Middle vec4
var Highlight vec4
var Phase float
var Spread float
var Fraction float
var Angle float
func rgb(v vec3)vec3{
 l:=v.x+0.3963377774*v.y+0.2158037573*v.z
 m:=v.x-0.1055613458*v.y-0.0638541728*v.z
 s:=v.x-0.0894841775*v.y-1.2914855480*v.z
 l=l*l*l;m=m*m*m;s=s*s*s
 return vec3(4.0767416621*l-3.3077115913*m+0.2309699292*s,-1.2684380046*l+2.6097574011*m-0.3413193965*s,-0.0041960863*l-0.7034186147*m+1.7076147010*s)
}
func lab(v vec3)vec3{
 l:=pow(max(0.0,dot(v,vec3(.4122214708,.5363325363,.0514459929))),1.0/3.0)
 m:=pow(max(0.0,dot(v,vec3(.2119034982,.6806995451,.1073969566))),1.0/3.0)
 s:=pow(max(0.0,dot(v,vec3(.0883024619,.2817188376,.6299787005))),1.0/3.0)
 return vec3(.2104542553*l+.793617785*m-.0040720468*s,1.9779984951*l-2.428592205*m+.4505937099*s,.0259040371*l+.7827717662*m-.808675766*s)
}
func inGamut(v vec3)bool{return v.x>=0&&v.y>=0&&v.z>=0&&v.x<=1&&v.y<=1&&v.z<=1}
func encode(v float)float{if v<=.0031308{return v*12.92};return 1.055*pow(v,1.0/2.4)-.055}
// CSS Color 4's local-MINDE chroma reduction (DeltaE OK JND .02).
func gamut(v vec3)vec3{
 if v.x>=1{return vec3(1)};if v.x<=0{return vec3(0)}
 raw:=rgb(v);if inGamut(raw){return vec3(encode(raw.x),encode(raw.y),encode(raw.z))}
 chroma:=length(v.yz);if chroma<.000001{return vec3(encode(clamp(raw.x,0,1)),encode(clamp(raw.y,0,1)),encode(clamp(raw.z,0,1)))}
 low:=0.0;high:=chroma;lowInGamut:=true
 clipped:=clamp(raw,vec3(0),vec3(1))
 for i:=0;i<18;i++{
  if high-low<.0001{break}
  current:=(low+high)/2
  candidate:=vec3(v.x,v.yz*(current/chroma))
  raw=rgb(candidate)
  if lowInGamut&&inGamut(raw){low=current;continue}
  clipped=clamp(raw,vec3(0),vec3(1))
  delta:=length(lab(clipped)-candidate)
  if delta<.02{
   if .02-delta<.0001{break}
   lowInGamut=false;low=current
  }else{high=current}
 }
 return vec3(encode(clipped.x),encode(clipped.y),encode(clipped.z))
}
func blend(a vec4,b vec4,t float)vec4{
 alpha:=mix(a.a,b.a,t);if alpha<=0{return vec4(0)}
 return vec4((a.xyz*a.a*(1-t)+b.xyz*b.a*t)/alpha,alpha)
}
func ColorEffect(pos vec2,color vec4)vec4{
 if color.a<=0{return vec4(0)}
 size:=rectSize();spread:=Spread
 if Fraction>=0{spread=size.x*Fraction}
 width:=size.x*2+spread*2
 axis:=vec2(cos(Angle),sin(Angle))
 line:=abs(axis.x)*width+abs(axis.y)*size.y
 band:=spread;if Fraction>=0{band=line*Fraction}
 offset:=(size.x-width)*(1-Phase)
 distance:=dot(pos-vec2(offset+width/2,size.y/2),axis)
 sample:=Base
 if band>0&&distance>-band&&distance<band{
  if distance< -band*.5{sample=blend(Base,Middle,(distance+band)/(band*.5))
  }else if distance<0{sample=blend(Middle,Highlight,(distance+band*.5)/(band*.5))
  }else if distance<band*.5{sample=blend(Highlight,Middle,distance/(band*.5))
  }else{sample=blend(Middle,Base,(distance-band*.5)/(band*.5))}
 }
 alpha:=sample.a*color.a
 return vec4(gamut(sample.xyz)*alpha,alpha)
}
`)
