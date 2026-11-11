package gallery

import (
	"encoding/json"
	"regexp"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func TestGalleryCatalogsAreCompleteAndValid(t *testing.T) {
	var english, chinese map[string]string
	placeholders := regexp.MustCompile(`%%|%\d*(?:\.\d+)?[dsqfxX]`)
	for _, catalog := range []struct {
		locale string
		values *map[string]string
	}{{LocaleEnglish, &english}, {LocaleChinese, &chinese}} {
		data, err := localeFiles.ReadFile("locales/" + catalog.locale + ".json")
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, catalog.values); err != nil {
			t.Fatal(err)
		}
		for key, value := range *catalog.values {
			if key == "@@locale" {
				continue
			}
			_, locale, ok := messages.Message(catalog.locale, key)
			if !ok || locale != catalog.locale || value == "" {
				t.Errorf("invalid or missing %s message %q", catalog.locale, key)
			}
			if catalog.locale == LocaleEnglish && messages.Format(LocaleEnglish, key, nil) != key {
				t.Errorf("English source changed while formatting %q", key)
			}
		}
	}
	for key := range english {
		if translated, ok := chinese[key]; !ok {
			t.Errorf("missing Chinese translation for %q", key)
		} else if !slices.Equal(placeholders.FindAllString(key, -1), placeholders.FindAllString(translated, -1)) {
			t.Errorf("Chinese translation changed the formatting arguments for %q", key)
		}
	}
	for key := range chinese {
		if _, ok := english[key]; !ok {
			t.Errorf("missing English source for %q", key)
		}
	}
}

func TestGalleryLanguageSwitchPreservesDraftsAndBilingualSearch(t *testing.T) {
	s := New()
	s.Page = "controls"
	s.Config.NativeMenus = true
	r := queryStateTester(s)
	field := websiteQueryField(t, r)
	queryFieldClick(r, field)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("未保存的 draft")
	wantPreset, wantHistory := s.Config, len(s.undo)
	requested := ""
	s.OnLocaleChange = func(locale string) { requested = locale }
	if err := r.Click("Switch to Chinese"); err != nil {
		t.Fatal(err)
	}
	if requested != LocaleChinese {
		t.Fatalf("requested locale = %q", requested)
	}
	// A host applies this callback through its own locale setter; the native
	// Gallery uses App.SetLocale and headless hosts use Tester.SetLocale.
	r.SetLocale(requested)
	if _, ok := r.AccessByRole(ui.RoleTab, "表单与控件"); !ok || !r.HasText("设计系统") {
		t.Fatal("language switch did not update chrome and navigation")
	}
	if got := advancedAccess(t, r, ui.RoleTextField, "网站名称").Value; got != "未保存的 draft" {
		t.Fatalf("language switch changed the native input draft: %q", got)
	}
	if s.Config != wantPreset || len(s.undo) != wantHistory || s.Page != "controls" {
		t.Fatal("language switch changed the design, history or current route")
	}
	for _, query := range []string{"输入框", "Input", "Forms & controls"} {
		r.Key(ui.Cmd, ui.KeyK)
		r.Frame()
		r.Key(ui.Cmd, ui.KeyA)
		r.Type(query)
		r.Key(0, ui.KeyEnter)
		r.Frame()
		if s.Command || s.Page != "controls" {
			t.Fatalf("bilingual search %q did not navigate to Forms", query)
		}
		if query != "Forms & controls" {
			if _, ok := r.AccessByRole(ui.RoleMenuButton, "输入框配置"); !ok || s.pendingPreview != "" {
				t.Fatalf("search %q did not reveal the localized source configuration", query)
			}
		}
	}
	if err := r.Click("切换到英文"); err != nil {
		t.Fatal(err)
	}
	if requested != LocaleEnglish {
		t.Fatalf("requested return locale = %q", requested)
	}
	r.SetLocale(requested)
	if !r.HasText("Design system") || s.websiteName != "未保存的 draft" || s.Config != wantPreset {
		t.Fatal("switching back lost the draft, design or English text")
	}
}

func TestGalleryChinesePagesAndCompactLanguageControl(t *testing.T) {
	for _, dark := range []bool{false, true} {
		for _, page := range []struct{ key, title string }{
			{"overview", "创造美好的界面。"}, {"controls", "表单与控件"},
			{"display", "数据与展示"}, {"content", "让内容更有意义。"},
			{"text-layout", "文字、代码与空间。"}, {"overlays", "浮层"},
			{"conversation", "聊一聊。"}, {"navigation", "找到你的方向。"},
			{"themes", "协调一致的配色。"},
		} {
			t.Run(page.key+map[bool]string{false: "/light", true: "/dark"}[dark], func(t *testing.T) {
				s := New()
				s.Page = page.key
				s.Config.ReduceMotion = true
				if dark {
					s.Theme.Mode = theme.DarkMode
				}
				r := ui.NewTester(s.View, 1280, 900)
				r.SetLocale("zh-CN")
				if !r.HasText(page.title) {
					t.Fatalf("Chinese page title missing: %s", page.title)
				}
			})
		}
	}
	s := New()
	r := ui.NewTester(s.View, 500, 900)
	for _, locale := range []string{LocaleEnglish, LocaleChinese} {
		r.SetLocale(locale)
		label := "Switch to Chinese"
		if locale == LocaleChinese {
			label = "切换到英文"
		}
		button := advancedAccess(t, r, ui.RoleButton, label)
		if button.Offscreen || button.Bounds.X < 0 || button.Bounds.X+button.Bounds.W > 500 {
			t.Fatalf("compact language button is clipped: %+v", button.Bounds)
		}
	}
}

func TestGalleryLocalizedChoicesKeepStableValuesAndSourceLabels(t *testing.T) {
	s := New()
	s.Config.NativeMenus = true
	r := ui.NewTester(s.View, 1280, 1000)
	r.SetLocale(LocaleChinese)
	field := advancedAccess(t, r, ui.RoleMenuButton, "外观")
	queryFieldClick(r, field)
	if err := r.ChooseMenuItem("深色"); err != nil {
		t.Fatal(err)
	}
	if s.Config.Appearance != theme.DarkMode {
		t.Fatalf("translated appearance published a translated identity: %q", s.Config.Appearance)
	}
	if !s.Undo() || s.Config.Appearance != theme.System {
		t.Fatal("localized choices broke design undo")
	}
	english := ui.NewTester(New().View, 1280, 900)
	if !english.HasText("Design system") || cardCompositionRecipes[0].Label != "Default size" {
		t.Fatal("Chinese rendering mutated shared English data")
	}
}

func TestGalleryChartCatalogueLocalizesControlsAndKeepsInspection(t *testing.T) {
	s := New()
	r := ui.NewTester(s.chartSourcePreview, 800, 1000)
	r.SetLocale(LocaleChinese)
	if !r.HasText("图表源目录") {
		t.Fatal("embedded chart catalogue did not inherit the Gallery language")
	}
	field := advancedAccess(t, r, ui.RoleTextField, "查找图表示例")
	queryFieldClick(r, field)
	r.Type("交互")
	if s.chartCatalogue.ID != "chart-area-interactive" {
		t.Fatalf("localized chart search selected %q", s.chartCatalogue.ID)
	}
	r.Key(ui.Cmd, ui.KeyA)
	r.Type("chart-bar-stacked")
	model := s.chartCatalogue.Model(s.chartCatalogue.ID)
	model.Selected = 2
	wantID := s.chartCatalogue.ID
	r.SetLocale(LocaleEnglish)
	if !r.HasText("Chart source catalogue") || s.chartCatalogue.Query != "chart-bar-stacked" || s.chartCatalogue.ID != wantID || model.Selected != 2 {
		t.Fatal("switching catalogue language lost the search or inspection state")
	}
}
