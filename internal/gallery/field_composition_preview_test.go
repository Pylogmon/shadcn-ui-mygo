package gallery

import (
	"image/png"
	"math"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func fieldPreviewAccess(t *testing.T, r *ui.Tester, role ui.Role, label string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, label)
	if !ok || a.Bounds.W <= 0 || a.Bounds.H <= 0 {
		t.Fatalf("missing Field control %q role %v", label, role)
	}
	return a
}
func fieldPreviewClick(t *testing.T, r *ui.Tester, label string) {
	t.Helper()
	a, ok := r.Access(label)
	if !ok || a.Offscreen || a.Bounds.W <= 0 || a.Bounds.H <= 0 {
		t.Fatalf("missing visible Field action %q: %+v", label, a)
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}
func fieldPreviewType(t *testing.T, r *ui.Tester, label, value string) {
	t.Helper()
	fieldPreviewClick(t, r, label)
	r.Key(ui.Cmd, ui.KeyA)
	r.Type(value)
	r.Frame()
}
func fieldPreviewSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_FIELD_COMPOSITION_SNAPSHOT_DIR")
	if dir == "" {
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(dir, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	err = png.Encode(f, r.Image())
	closeErr := f.Close()
	if err != nil || closeErr != nil {
		t.Fatalf("Field snapshot %v/%v", err, closeErr)
	}
}

// One local application workflow uses the real native editors/choices while
// switching source compositions. Whole Gallery routing is tested by its owner.
func TestFieldCompositionSourceWorkflow(t *testing.T) {
	p := &fieldCompositionPreviewState{}
	backend := &creatorPopupBackend{}
	host := desktop.NewHostWithBackend(backend)
	dark, native, visible := false, true, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Fill().Padding(20).Gap(12).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			host.BeginFrame()
			defer host.EndFrame()
			ui.Row(c).Gap(8).Children(func() {
				if ui.Button(c, "Leave field examples").Clicked() {
					visible = false
				}
				if ui.Button(c, "Show field examples").Clicked() {
					visible = true
				}
			})
			if visible {
				p.view(c, host, native)
			} else {
				ui.Text(c, "Another local page")
			}
		})
	}, 620, 1120, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	choose := func(label string, items ...string) {
		t.Helper()
		a := fieldPreviewAccess(t, r, ui.RoleMenuButton, label)
		r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
		r.Frame()
		if err := r.ChooseMenuItem(items...); err != nil {
			t.Fatal(err)
		}
		r.Frame()
	}
	configure := func(item string) { t.Helper(); choose("Field configuration", item) }
	variant := func(item string) { t.Helper(); choose("Field variant", item) }
	if p.recipe != "responsive" || p.popupError != "" {
		t.Fatal("initial field source recipe or quiet Host contract changed")
	}
	fieldPreviewClick(t, r, "Submit profile")
	if p.profileError != "Enter your name" || !fieldPreviewAccess(t, r, ui.RoleTextField, "Name *").Invalid {
		t.Fatal("required profile did not validate its actual editor")
	}
	fieldPreviewType(t, r, "Name *", "Ada Lovelace")
	fieldPreviewClick(t, r, "Submit profile")
	if p.profileSaved != "Ada Lovelace" || p.profileError != "" {
		t.Fatal("profile submit did not save local application state")
	}
	name := fieldPreviewAccess(t, r, ui.RoleTextField, "Name *")
	description, ok := r.Access("Provide your full name for identification")
	if !ok || description.Bounds.X >= name.Bounds.X {
		t.Fatal("wide responsive Field did not keep label/description beside the editor")
	}
	r.SetSize(390, 1120)
	r.Frame()
	name = fieldPreviewAccess(t, r, ui.RoleTextField, "Name *")
	description, ok = r.Access("Provide your full name for identification")
	if !ok || description.Bounds.Y+description.Bounds.H > name.Bounds.Y+1 {
		t.Fatalf("narrow FieldGroup did not stack its responsive content: helper=%+v editor=%+v", description, name)
	}
	fieldPreviewSnapshot(t, r, "field-profile-narrow-light")
	r.SetSize(620, 1120)
	r.Frame()
	configure("Username and password")
	fieldPreviewType(t, r, "Username", "native-go")
	fieldPreviewType(t, r, "Password", "local-password")
	password := fieldPreviewAccess(t, r, ui.RoleTextField, "Password")
	helper, ok := r.Access("Must be at least 8 characters long.")
	if !ok || helper.Bounds.Y+helper.Bounds.H > password.Bounds.Y+1 || !slices.Contains(password.DescribedBy, "Must be at least 8 characters long.") {
		t.Fatal("password description was not before and related to its actual editor")
	}
	configure("Compute environment")
	if *p.text("compute:environment", "") != "kubernetes" {
		t.Fatal("source compute default is not Kubernetes")
	}
	vm := fieldPreviewAccess(t, r, ui.RoleRadio, "Virtual Machine")
	caption, ok := r.Access("Access a cluster to run GPU workloads.")
	if !ok {
		t.Fatal("choice card description missing")
	}
	r.ClickAt(caption.Bounds.X+4, caption.Bounds.Y+caption.Bounds.H/2)
	r.Frame()
	if *p.text("compute:environment", "") != "vm" || !fieldPreviewAccess(t, r, ui.RoleRadio, "Virtual Machine").Checked {
		t.Fatalf("clicking the card description did not activate the actual radio: %+v", vm)
	}
	dark = true
	r.SetSize(420, 1120)
	r.SetDirection(ui.RightToLeft)
	r.Frame()
	fieldPreviewSnapshot(t, r, "field-compute-dark-rtl")
	r.SetDirection(ui.LeftToRight)
	r.SetSize(620, 1120)
	dark = false
	r.Frame()
	configure("Payment fields")
	fieldPreviewType(t, r, "Name on Card", "Local Demo")
	fieldPreviewType(t, r, "Card Number", "1111 2222 3333 4444")
	choose("Month", "03")
	choose("Year", "2028")
	fieldPreviewType(t, r, "CVV", "123")
	fieldPreviewType(t, r, "Comments", "Keep this example local")
	fieldPreviewClick(t, r, "Same as shipping address")
	fieldPreviewClick(t, r, "Submit payment fields")
	saved := p.paymentSaved["payment"]
	if saved.Name != "Local Demo" || saved.Month != "03" || saved.Year != "2028" || saved.Shipping || saved.Comments != "Keep this example local" {
		t.Fatalf("fieldset controls did not reach the saved local example: %+v", saved)
	}
	fieldPreviewType(t, r, "Name on Card", "Changed draft")
	fieldPreviewClick(t, r, "Cancel payment fields")
	if *p.text("payment:card-name", "") != "Local Demo" {
		t.Fatal("payment cancel failed to restore the independent snapshot")
	}
	fieldPreviewSnapshot(t, r, "field-payment-light")
	configure("Create: Input fields")
	variant("Input with Badge")
	if !r.HasText("Recommended") {
		t.Fatal("rich field label lost the source Badge")
	}
	fieldPreviewType(t, r, "Input with Badge", "badge draft")
	variant("Invalid Input")
	if !fieldPreviewAccess(t, r, ui.RoleTextField, "Invalid Input").Invalid {
		t.Fatal("source invalid input did not reach the real editor")
	}
	configure("Create: OTP fields")
	variant("PIN Code")
	fieldPreviewType(t, r, "PIN Code", "1a2345")
	if *p.text("create-otp:pin", "") != "1234" {
		t.Fatal("four-digit native PIN did not preserve its source input contract")
	}
	configure("Create: Native select fields")
	variant("Grouped Options")
	choose("Grouped Options", "Germany")
	if *p.text("create-native:grouped", "") != "de" {
		t.Fatal("grouped native select did not publish the actual selected value")
	}
	configure("Create: Slider fields")
	variant("Temperature Range")
	low := fieldPreviewAccess(t, r, ui.RoleSlider, "Temperature Range minimum")
	r.ClickAt(low.Bounds.X+low.Bounds.W/2, low.Bounds.Y+low.Bounds.H/2)
	r.Key(0, ui.KeyRight)
	r.Frame()
	if (*p.values("create-slider:temperature", .3, .7))[0] <= .3 || !r.HasText("Range: 0.4 - 0.7") {
		t.Fatal("native range keyboard input did not update the field description")
	}
	configure("Create: Horizontal fields")
	variant("Email Notifications")
	fieldPreviewClick(t, r, "Email Notifications")
	if !*p.checked("create-horizontal:notifications", false) {
		t.Fatal("horizontal content label did not activate its Switch")
	}
	// These two source groups have distinct 12DIP composition contracts.
	// Compare complete caption rows (checkboxes are shorter than their labels),
	// and actual Switch rows so their larger glyph does not skew the gap.
	configure("Create: Checkbox fields")
	variant("Preferences")
	preferenceLabels := []string{"Dark mode", "Compact view", "Enable notifications"}
	for index := 1; index < len(preferenceLabels); index++ {
		previous := fieldPreviewAccess(t, r, ui.RoleText, preferenceLabels[index-1])
		next := fieldPreviewAccess(t, r, ui.RoleText, preferenceLabels[index])
		gap := next.Bounds.Y - previous.Bounds.Y - previous.Bounds.H
		if math.Abs(float64(gap-12)) > .1 {
			t.Errorf("source Preferences row gap = %.2f, want 12DIP", gap)
		}
	}
	configure("Create: Switch fields")
	variant("Privacy Settings")
	privacyTitle := fieldPreviewAccess(t, r, ui.RoleText, "Privacy Settings")
	privacyDescription := fieldPreviewAccess(t, r, ui.RoleText, "Manage your privacy preferences.")
	profileSwitch := fieldPreviewAccess(t, r, ui.RoleSwitch, "Make profile visible to others")
	emailSwitch := fieldPreviewAccess(t, r, ui.RoleSwitch, "Show email on profile")
	profileCaption := fieldPreviewAccess(t, r, ui.RoleText, "Make profile visible to others")
	emailCaption := fieldPreviewAccess(t, r, ui.RoleText, "Show email on profile")
	if profileSwitch.Bounds.Y < profileCaption.Bounds.Y || profileSwitch.Bounds.Y+profileSwitch.Bounds.H > profileCaption.Bounds.Y+profileCaption.Bounds.H+.1 || emailSwitch.Bounds.Y < emailCaption.Bounds.Y || emailSwitch.Bounds.Y+emailSwitch.Bounds.H > emailCaption.Bounds.Y+emailCaption.Bounds.H+.1 {
		t.Error("Privacy Switch glyph escaped its complete caption row")
	}
	privacyGaps := []float32{
		privacyDescription.Bounds.Y - privacyTitle.Bounds.Y - privacyTitle.Bounds.H,
		profileCaption.Bounds.Y - privacyDescription.Bounds.Y - privacyDescription.Bounds.H,
		emailCaption.Bounds.Y - profileCaption.Bounds.Y - profileCaption.Bounds.H,
	}
	for index, gap := range privacyGaps {
		if math.Abs(float64(gap-12)) > .1 {
			t.Errorf("source Privacy part %d gap = %.2f, want 12DIP", index, gap)
		}
	}
	privacyLabel := fieldPreviewAccess(t, r, ui.RoleGroup, "Privacy Settings")
	if len(privacyLabel.DescribedBy) != 0 {
		t.Error("source Privacy label was replaced by a described FieldSet")
	}
	configure("Translated payment")
	choose("Field language", "עברית")
	if !r.HasText("אמצעי תשלום") || !r.HasText("כתובת חיוב") {
		t.Fatal("fixed RTL field translations were not used")
	}
	dark = true
	r.SetSize(420, 1120)
	r.Frame()
	fieldPreviewSnapshot(t, r, "field-payment-hebrew-dark")
	fieldPreviewClick(t, r, "Leave field examples")
	fieldPreviewClick(t, r, "Show field examples")
	configure("Responsive profile")
	if fieldPreviewAccess(t, r, ui.RoleTextField, "Name *").Value != "Ada Lovelace" || *p.text("input:username", "") != "native-go" || *p.text("create-input:badge", "") != "badge draft" {
		t.Fatal("source recipe/page changes lost independent application drafts")
	}
}
