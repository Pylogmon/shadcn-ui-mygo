#!/bin/sh
set -eu
if [ "$(uname -s)" != Darwin ]; then
    echo 'macOS app packaging requires macOS.' >&2
    exit 1
fi
cd "$(dirname "$0")/.."
case "${1:-gallery}" in
    gallery)
        bundle='build/Shadcn Gallery.app'
        package='./cmd/gallery'
        identifier='dev.pylogmon.shadcn-gallery'
        app_name='Shadcn Gallery'
        ;;
    native-popups)
        bundle='build/Shadcn Popups.app'
        package='./examples/native-popups'
        identifier='dev.pylogmon.shadcn-popups'
        app_name='Shadcn Popups'
        ;;
    chart-composition)
        bundle='build/Shadcn Charts.app'
        package='./examples/chart-composition'
        identifier='dev.pylogmon.shadcn-charts'
        app_name='Shadcn Charts'
        ;;
    chart-cartesian)
        bundle='build/Shadcn Cartesian.app'
        package='./examples/chart-cartesian'
        identifier='dev.pylogmon.shadcn-cartesian'
        app_name='Shadcn Cartesian'
        ;;
    chart-polar)
        bundle='build/Shadcn Polar.app'
        package='./examples/chart-polar'
        identifier='dev.pylogmon.shadcn-polar'
        app_name='Shadcn Polar'
        ;;
    chart-catalog)
        bundle='build/Shadcn Catalogue.app'
        package='./examples/chart-catalog'
        identifier='dev.pylogmon.shadcn-catalogue'
        app_name='Shadcn Catalogue'
        ;;
    tooltip-composition)
        bundle='build/Shadcn Tooltips.app'
        package='./examples/tooltip-composition'
        identifier='dev.pylogmon.shadcn-tooltips'
        app_name='Shadcn Tooltips'
        ;;
    tabs-composition)
        bundle='build/Shadcn Tabs.app'
        package='./examples/tabs-composition'
        identifier='dev.pylogmon.shadcn-tabs'
        app_name='Shadcn Tabs'
        ;;
    select-composition)
        bundle='build/Shadcn Selects.app'
        package='./examples/select-composition'
        identifier='dev.pylogmon.shadcn-selects'
        app_name='Shadcn Selects'
        ;;
    desktop-combobox)
        bundle='build/Shadcn Comboboxes.app'
        package='./examples/desktop-combobox'
        identifier='dev.pylogmon.shadcn-comboboxes'
        app_name='Shadcn Comboboxes'
        ;;
    typed-input)
        bundle='build/Shadcn Inputs.app'
        package='./examples/typed-input'
        identifier='dev.pylogmon.shadcn-inputs'
        app_name='Shadcn Inputs'
        ;;
    date-composition)
        bundle='build/Shadcn Dates.app'
        package='./examples/date-composition'
        identifier='dev.pylogmon.shadcn-dates'
        app_name='Shadcn Dates'
        ;;
    material-icons)
        bundle='build/Shadcn Symbols.app'
        package='./examples/material-icons'
        identifier='dev.pylogmon.shadcn-symbols'
        app_name='Shadcn Symbols'
        ;;
    *) echo 'Expected gallery, native-popups, chart-composition, chart-cartesian, chart-polar, chart-catalog, tooltip-composition, tabs-composition, select-composition, desktop-combobox, typed-input, date-composition or material-icons.' >&2; exit 1 ;;
esac
mkdir -p "$bundle/Contents/MacOS"
mkdir -p "$bundle/Contents/Resources/licenses"
cp pkg/overlay/testdata/cmdk/LICENSE.md "$bundle/Contents/Resources/licenses/cmdk.txt"
cp pkg/control/testdata/input-otp/LICENSE "$bundle/Contents/Resources/licenses/input-otp.txt"
cp internal/indicator/LICENSE.lucide "$bundle/Contents/Resources/licenses/lucide.txt"
cp pkg/overlay/testdata/sonner/LICENSE.md "$bundle/Contents/Resources/licenses/sonner.txt"
cp internal/shimmer/testdata/shadcn/LICENSE "$bundle/Contents/Resources/licenses/shadcn.txt"
cp internal/gallery/card-assets/README.md "$bundle/Contents/Resources/licenses/card-example-images.txt"
cp internal/gallery/item-assets/README.md "$bundle/Contents/Resources/licenses/item-example-images.txt"
cp pkg/theme/sources/LICENSE.tailwind "$bundle/Contents/Resources/licenses/tailwind.txt"
cp pkg/display/testdata/recharts-animation/LICENSE "$bundle/Contents/Resources/licenses/recharts.txt"
cp pkg/display/testdata/shadcn-chart/cartesian/LICENSE.d3-shape "$bundle/Contents/Resources/licenses/d3-shape.txt"
cp pkg/display/testdata/shadcn-chart/cartesian/LICENSE.d3-path "$bundle/Contents/Resources/licenses/d3-path.txt"
cp pkg/display/testdata/shadcn-carousel/npm/LICENSE.embla "$bundle/Contents/Resources/licenses/embla.txt"
cp pkg/desktop/testdata/shadcn-menu/npm/base-ui-react/LICENSE "$bundle/Contents/Resources/licenses/base-ui.txt"
cp pkg/desktop/testdata/shadcn-menu/npm/radix-ui-react-menu/LICENSE "$bundle/Contents/Resources/licenses/radix-ui.txt"
cp pkg/desktop/testdata/shadcn-menu/npm/tw-animate-css/LICENSE "$bundle/Contents/Resources/licenses/tw-animate-css.txt"
cp pkg/icons/material/sources/LICENSE "$bundle/Contents/Resources/licenses/material-symbols.txt"
cp pkg/display/testdata/shadcn-calendar/npm/date-fns-jalali/LICENSE.md "$bundle/Contents/Resources/licenses/date-fns-jalali.txt"
cp pkg/display/testdata/shadcn-calendar/npm/react-day-picker/LICENSE "$bundle/Contents/Resources/licenses/react-day-picker.txt"
# Font binaries are embedded by Go. Preserve every original text notice beside
# its family ID, including the notices for any supplemental static faces.
for font_dir in pkg/fonts/assets/*; do
    [ -d "$font_dir" ] || continue
    font_id=$(basename "$font_dir")
    for notice in "$font_dir"/*.txt; do
        [ -f "$notice" ] || continue
        mkdir -p "$bundle/Contents/Resources/licenses/fonts/$font_id"
        cp "$notice" "$bundle/Contents/Resources/licenses/fonts/$font_id/"
    done
done
CGO_ENABLED=0 go build -o "$bundle/Contents/MacOS/gallery" "$package"
cat > "$bundle/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
    <key>CFBundleExecutable</key><string>gallery</string>
    <key>CFBundleIdentifier</key><string>$identifier</string>
    <key>CFBundleName</key><string>$app_name</string>
    <key>CFBundlePackageType</key><string>APPL</string>
    <key>CFBundleShortVersionString</key><string>0.1.0</string>
    <key>NSHighResolutionCapable</key><true/>
</dict></plist>
PLIST
