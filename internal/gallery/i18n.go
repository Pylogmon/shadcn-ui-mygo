package gallery

import (
	"embed"
	"fmt"
	"io/fs"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/i18n"
	_ "github.com/egoist/mygo/locales/zh-Hans"
	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/control"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/conversation"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/display"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/overlay"
)

const (
	LocaleEnglish = "en"
	LocaleChinese = "zh-Hans"
)

// English source messages are also their keys. Untranslated technical names,
// code samples and user-entered text keep their original spelling.
//
//go:embed locales/*.json
var localeFiles embed.FS

var messages = func() *i18n.Catalog {
	files, err := fs.Sub(localeFiles, "locales")
	if err != nil {
		panic(err)
	}
	catalog, err := i18n.NewCatalog(files, LocaleEnglish)
	if err != nil {
		panic(err)
	}
	i18n.Register(catalog)
	return catalog
}()

func tr(c *ui.Context, source string) string {
	return messages.Format(c.Locale(), source, nil)
}

func trf(c *ui.Context, source string, args ...any) string {
	return fmt.Sprintf(tr(c, source), args...)
}

func localizedOptions(c *ui.Context, options []control.Option) []control.Option {
	if !strings.HasPrefix(c.Locale(), "zh") {
		return options
	}
	options = slices.Clone(options)
	for i := range options {
		options[i].Label = tr(c, options[i].Label)
		options[i].Description = tr(c, options[i].Description)
		options[i].Group = tr(c, options[i].Group)
	}
	return options
}

func localizedChart(c *ui.Context, props display.ChartProps) display.ChartProps {
	props.Labels = slices.Clone(props.Labels)
	for i := range props.Labels {
		props.Labels[i] = tr(c, props.Labels[i])
	}
	props.Series = slices.Clone(props.Series)
	for i := range props.Series {
		props.Series[i].Name = tr(c, props.Series[i].Name)
	}
	return props
}

func localizedMenuItems(c *ui.Context, items []overlay.MenuItem) []overlay.MenuItem {
	items = slices.Clone(items)
	for i := range items {
		items[i].Label = tr(c, items[i].Label)
		items[i].Group = tr(c, items[i].Group)
		items[i].Children = localizedMenuItems(c, items[i].Children)
	}
	return items
}

func localizedMenubarItems(c *ui.Context, items []overlay.MenubarItem) []overlay.MenubarItem {
	items = slices.Clone(items)
	for i := range items {
		items[i].Label = tr(c, items[i].Label)
		items[i].Items = localizedMenuItems(c, items[i].Items)
	}
	return items
}

func localizedQuestions(c *ui.Context, questions []conversation.Question) []conversation.Question {
	questions = slices.Clone(questions)
	for i := range questions {
		questions[i].Title = tr(c, questions[i].Title)
		questions[i].Description = tr(c, questions[i].Description)
		questions[i].InputLabel = tr(c, questions[i].InputLabel)
		questions[i].Placeholder = tr(c, questions[i].Placeholder)
		questions[i].Choices = localizedOptions(c, questions[i].Choices)
	}
	return questions
}

func localizedQuestionnaireProgress(c *ui.Context, parts *conversation.QuestionnaireParts, label string) {
	progress := parts.Progress(conversation.QuestionnaireProgressProps{
		Label: tr(c, label),
		Content: func(child *ui.Context, state conversation.QuestionnaireSnapshot) {
			if state.Total > 0 {
				ui.Text(child, trf(child, "Question %d of %d", state.Current, state.Total))
			}
		},
	})
	if parts.State.Total > 0 {
		progress.Value(trf(c, "Question %d of %d", parts.State.Current, parts.State.Total))
	}
}

func (s *State) languageSwitch(c *ui.Context) {
	locale, text, label := LocaleChinese, "中文", "Switch to Chinese"
	if strings.HasPrefix(c.Locale(), "zh") {
		locale, text, label = LocaleEnglish, "English", "Switch to English"
	}
	if control.Button(c.Key("gallery-language"), text, control.ButtonProps{Variant: control.Ghost, Size: control.Small, Label: tr(c, label)}).Clicked() {
		if s.OnLocaleChange != nil {
			s.OnLocaleChange(locale)
		} else {
			mygo.App.SetLocale(locale)
		}
	}
}
