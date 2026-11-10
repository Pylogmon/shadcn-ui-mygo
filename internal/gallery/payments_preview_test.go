package gallery

import (
	"errors"
	"image/png"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/desktop"
	"github.com/pylogmon/shadcn-ui-mygo/pkg/theme"
)

func paymentsAccess(t *testing.T, r *ui.Tester, role ui.Role, name string) ui.AccessInfo {
	t.Helper()
	a, ok := r.AccessByRole(role, name)
	if !ok || a.Offscreen {
		t.Fatalf("missing visible native control %v %q", role, name)
	}
	return a
}
func paymentsClick(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	a, ok := r.Access(name)
	if !ok || a.Offscreen {
		t.Fatal("missing payment action", name)
	}
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Frame()
}
func paymentsFilter(t *testing.T, r *ui.Tester, text string) {
	t.Helper()
	a := paymentsAccess(t, r, ui.RoleTextField, "Filter payment emails")
	r.ClickAt(a.Bounds.X+a.Bounds.W/2, a.Bounds.Y+a.Bounds.H/2)
	r.Key(ui.Cmd, ui.KeyA)
	if text == "" {
		r.Key(0, ui.KeyBackspace)
	} else {
		r.Type(text)
	}
	r.Frame()
}
func paymentsEmails(r *ui.Tester) []string {
	var emails []string
	for _, a := range r.AccessAll() {
		if a.Role == ui.RoleCell && strings.Contains(a.Name, "@") {
			emails = append(emails, a.Name)
		}
	}
	return emails
}
func paymentsChoose(t *testing.T, r *ui.Tester, b *creatorPopupBackend, native bool, trigger, choice string) {
	t.Helper()
	count := len(b.surfaces)
	paymentsClick(t, r, trigger)
	if native {
		if count != len(b.surfaces) || !slices.Contains(r.Menu(), choice) {
			t.Fatal("expected actual OS payment menu")
		}
		if err := r.ChooseMenuItem(choice); err != nil {
			t.Fatal(err)
		}
	} else {
		if len(b.surfaces) != count+1 {
			t.Fatal("expected independent payment Flyout")
		}
		child := b.surfaces[len(b.surfaces)-1]
		child.tester.Advance(time.Second)
		if err := child.tester.Click(choice); err != nil {
			t.Fatal(err)
		}
	}
	r.Frame()
}
func paymentsSnapshot(t *testing.T, r *ui.Tester, name string) {
	t.Helper()
	dir := os.Getenv("GALLERY_PAYMENTS_SNAPSHOT_DIR")
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
		t.Fatalf("snapshot %v/%v", err, closeErr)
	}
}

func TestPaymentsPreviewSourceWorkflow(t *testing.T) {
	p := &paymentsPreviewState{}
	b := &creatorPopupBackend{}
	h := desktop.NewHostWithBackend(b)
	native, dark, show := false, false, true
	r := ui.NewTesterWith(func(c *ui.Context) {
		theme.Apply(c)
		h.BeginFrame()
		defer h.EndFrame()
		tokens := theme.Preset("blue", dark)
		theme.Scope(c, tokens).Column().Fill().Padding(24).Background(tokens.Background).TextColor(tokens.Foreground).Children(func() {
			if show {
				p.view(c, h, native)
			} else {
				ui.Text(c, "Another Gallery page")
			}
		})
	}, 1040, 900, ui.TesterOptions{Preferences: ui.Preferences{ReduceMotion: true}})
	if p.table.PageSize != 10 || len(paymentsEmails(r)) != 5 || !paymentsAccess(t, r, ui.RoleButton, "Next payments page").Disabled || p.popupError != "" {
		t.Fatal("source five-row/default ten-row page was changed or reported an inactive error")
	}
	paymentsFilter(t, r, "abe45")
	if !slices.Equal(paymentsEmails(r), []string{"abe45@example.com"}) {
		t.Fatal("email filter did not bind real editor to table rows")
	}
	paymentsFilter(t, r, "")
	paymentsClick(t, r, "Sort payment emails")
	if !slices.Equal(paymentsEmails(r), []string{"abe45@example.com", "carmella@example.com", "ken99@example.com", "monserrat44@example.com", "silas22@example.com"}) || paymentsAccess(t, r, ui.RoleColumnHeader, "Email").Sort != ui.SortAscending {
		t.Fatal("source email ascending order/AX sort missing")
	}
	paymentsClick(t, r, "Sort payment emails")
	if paymentsEmails(r)[0] != "silas22@example.com" || paymentsAccess(t, r, ui.RoleColumnHeader, "Email").Sort != ui.SortDescending {
		t.Fatal("source email descending order/AX sort missing")
	}
	paymentsClick(t, r, "Sort payment emails")
	paymentsChoose(t, r, b, native, "Payments page size", "2 per page (demo)")
	if p.table.PageSize != 2 || len(paymentsEmails(r)) != 2 || paymentsAccess(t, r, ui.RoleButton, "Next payments page").Disabled {
		t.Fatal("explicit small demo page size did not enable actual paging")
	}
	paymentsClick(t, r, "Select payment 3u1reuv4")
	header := paymentsAccess(t, r, ui.RoleCheckBox, "Select all payments on page")
	if !header.Mixed || header.Checked {
		t.Fatal("partial page did not produce native mixed checkbox")
	}
	paymentsClick(t, r, "Select all payments on page")
	if p.table.Selection.Len() != 2 || !paymentsAccess(t, r, ui.RoleCheckBox, "Select all payments on page").Checked {
		t.Fatal("mixed activation did not select all current-page rows")
	}
	paymentsClick(t, r, "Select all payments on page")
	if p.table.Selection.Len() != 0 || paymentsAccess(t, r, ui.RoleCheckBox, "Select all payments on page").Checked {
		t.Fatal("all-page toggle did not clear page")
	}
	paymentsClick(t, r, "Select payment 3u1reuv4")
	paymentsClick(t, r, "Next payments page")
	paymentsClick(t, r, "Select payment derv1ws0")
	if !p.table.Selection.Has("3u1reuv4") || !p.table.Selection.Has("derv1ws0") || !r.HasText("2 of 5 row(s) selected.") {
		t.Fatal("pagination lost stable source keys/filtered selection count")
	}
	paymentsClick(t, r, "Payment columns")
	child := b.surfaces[len(b.surfaces)-1]
	child.tester.Advance(time.Second)
	columns := []string{}
	for _, a := range child.tester.AccessAll() {
		if a.Role == ui.RoleMenuItemCheckBox {
			columns = append(columns, a.Name)
		}
	}
	if !slices.Equal(columns, []string{"Status", "Email", "Amount"}) {
		t.Fatalf("protected selection/actions leaked into hideable columns: %v", columns)
	}
	if err := child.tester.Click("Amount"); err != nil {
		t.Fatal(err)
	}
	r.Frame()
	if !p.table.Hidden["amount"] {
		t.Fatal("column checkbox did not update app visibility")
	}
	if _, ok := r.AccessByRole(ui.RoleColumnHeader, "Amount"); ok {
		t.Fatal("hidden column retained native header")
	}
	paymentsAccess(t, r, ui.RoleCheckBox, "Select all payments on page")
	paymentsAccess(t, r, ui.RoleButton, "Payment actions m5gr84i9")
	// The row is on the sorted second page; copy must use its stable original ID,
	// and successful status must come from the real asynchronous clipboard result.
	r.SetClipboard("previous clipboard")
	r.FailClipboard(errors.New("test clipboard refusal"))
	paymentsChoose(t, r, b, native, "Payment actions m5gr84i9", "Copy payment ID")
	if r.Clipboard() != "previous clipboard" || p.copiedID != "" || !strings.Contains(p.feedback, "Could not copy payment ID m5gr84i9") {
		t.Fatal("failed payment copy reported success")
	}
	r.FailClipboard(nil)
	paymentsChoose(t, r, b, native, "Payment actions m5gr84i9", "Copy payment ID")
	if r.Clipboard() != "m5gr84i9" || p.copiedID != "m5gr84i9" || p.feedback != "Copied payment ID m5gr84i9." {
		t.Fatal("row clipboard action copied wrong payment or bypassed acknowledgment")
	}
	paymentsChoose(t, r, b, native, "Payment actions m5gr84i9", "View payment details")
	if p.detailID != "m5gr84i9" || !r.HasText("Payment details · m5gr84i9") {
		t.Fatal("local row-details action targeted wrong record")
	}
	paymentsSnapshot(t, r, "payments-light")
	native = true
	r.Frame()
	paymentsChoose(t, r, b, native, "Payment columns", "Amount")
	if p.table.Hidden["amount"] {
		t.Fatal("OS column menu did not restore column")
	}
	paymentsChoose(t, r, b, native, "Payments page size", "5 per page")
	if p.table.Page != 0 || p.table.PageSize != 5 || len(paymentsEmails(r)) != 5 || p.table.Selection.Len() != 2 {
		t.Fatal("page-size change reset selection or failed to clamp page")
	}
	dark = true
	r.SetDirection(ui.RightToLeft)
	r.Frame()
	paymentsSnapshot(t, r, "payments-dark-rtl")
	action, ok := r.Access("Payment actions 3u1reuv4")
	if !ok {
		t.Fatal("missing native payment row action")
	}
	if paymentsAccess(t, r, ui.RoleCheckBox, "Select payment 3u1reuv4").Bounds.X <= action.Bounds.X {
		t.Fatal("native table did not mirror selection/actions in RTL")
	}
	// Status is a real cell value but must not match the email-only filter.
	paymentsFilter(t, r, "success")
	if len(paymentsEmails(r)) != 0 || !r.HasText("No results.") || !r.HasText("0 of 0 row(s) selected.") || p.table.Selection.Len() != 2 {
		t.Fatal("email filter became global search or lost hidden selected keys")
	}
	empty := paymentsAccess(t, r, ui.RoleCell, "No results.")
	if empty.ColumnSpan != 5 || !paymentsAccess(t, r, ui.RoleCheckBox, "Select all payments on page").Disabled || !paymentsAccess(t, r, ui.RoleButton, "Next payments page").Disabled || !paymentsAccess(t, r, ui.RoleButton, "Previous payments page").Disabled {
		t.Fatal("empty-state span/header/pager invalid")
	}
	paymentsFilter(t, r, "")
	show = false
	r.Frame()
	show = true
	r.Frame()
	if p.table.PageSize != 5 || p.table.Sort.Column != "email" || p.table.Sort.Descending || p.table.Selection.Len() != 2 || p.copiedID != "m5gr84i9" || len(paymentsEmails(r)) != 5 {
		t.Fatal("page remount lost app-owned filter/sort/selection/copy state")
	}
	paymentsAccess(t, r, ui.RoleTable, "Recent payments")
}
