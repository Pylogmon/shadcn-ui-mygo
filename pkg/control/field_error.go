package control

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

// FieldErrorProps supplies validator messages or custom native error content.
// Empty messages are omitted and exact duplicates keep their first position.
// Content takes precedence over the default text/list. Its accessible Label
// defaults to the joined Errors; provide Label when Content has no Errors.
type FieldErrorProps struct {
	Errors  []string
	Label   string
	Content func(*ui.Context)
}

type fieldErrorContent struct {
	messages []string
	label    string
	build    func(*ui.Context)
}

func resolveFieldError(p FieldErrorProps) fieldErrorContent {
	content := fieldErrorContent{label: p.Label, build: p.Content}
	seen := make(map[string]bool, len(p.Errors))
	for _, message := range p.Errors {
		if message != "" && !seen[message] {
			seen[message] = true
			content.messages = append(content.messages, message)
		}
	}
	if content.label == "" {
		content.label = strings.Join(content.messages, "\n")
	}
	if content.build != nil && content.label == "" {
		panic("control.FieldError: custom content needs Label or nonempty Errors")
	}
	return content
}

func (p fieldErrorContent) present() bool { return p.build != nil || len(p.messages) != 0 }

// FieldError renders a native alert: one message as text, multiple messages as
// a semantic bulleted list. With no messages/content it returns an invalid
// Element and declares no alert. Callers can use ErrorMessage/DescribedBy to
// associate this independently composed element with their editor.
func FieldError(c *ui.Context, p FieldErrorProps) ui.Element {
	return renderFieldError(c, resolveFieldError(p))
}

func renderFieldError(c *ui.Context, p fieldErrorContent) ui.Element {
	if !p.present() {
		return ui.Element{}
	}
	color := theme.TokensOf(c).Destructive
	if p.build == nil && len(p.messages) == 1 {
		return ui.Text(c, p.messages[0]).MinWidth(0).FontSize(14).LineHeight(20.0 / 14).FontWeight(400).TextColor(color).Role(ui.RoleAlert).Label(p.label)
	}
	return ui.Column(c).FillWidth().MinWidth(0).FontSize(14).LineHeight(20.0 / 14).FontWeight(400).TextColor(color).Role(ui.RoleAlert).Label(p.label).Children(func() {
		if p.build != nil {
			p.build(c)
			return
		}
		ui.Column(c).FillWidth().MinWidth(0).Gap(4).PaddingStart(16).Role(ui.RoleList).Label("Validation errors").Children(func() {
			for index, message := range p.messages {
				ui.Row(c.Key(message)).FillWidth().MinWidth(0).Gap(6).AlignItems(ui.Start).Role(ui.RoleListItem).Label(message).PositionInSet(index+1, len(p.messages)).Children(func() {
					ui.Text(c, "•").Shrink(0).AccessHidden(true)
					ui.Text(c, message).MinWidth(0).Grow(1).AccessHidden(true)
				})
			}
		})
	})
}
