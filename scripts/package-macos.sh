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
    material-icons)
        bundle='build/Shadcn Symbols.app'
        package='./examples/material-icons'
        identifier='dev.pylogmon.shadcn-symbols'
        app_name='Shadcn Symbols'
        ;;
    *) echo 'Expected gallery, native-popups, chart-composition or material-icons.' >&2; exit 1 ;;
esac
mkdir -p "$bundle/Contents/MacOS"
mkdir -p "$bundle/Contents/Resources/licenses"
cp pkg/overlay/testdata/cmdk/LICENSE.md "$bundle/Contents/Resources/licenses/cmdk.txt"
cp pkg/control/testdata/input-otp/LICENSE "$bundle/Contents/Resources/licenses/input-otp.txt"
cp internal/indicator/LICENSE.lucide "$bundle/Contents/Resources/licenses/lucide.txt"
cp pkg/overlay/testdata/sonner/LICENSE.md "$bundle/Contents/Resources/licenses/sonner.txt"
cp internal/shimmer/testdata/shadcn/LICENSE "$bundle/Contents/Resources/licenses/shadcn.txt"
cp pkg/theme/sources/LICENSE.tailwind "$bundle/Contents/Resources/licenses/tailwind.txt"
cp pkg/display/testdata/recharts-animation/LICENSE "$bundle/Contents/Resources/licenses/recharts.txt"
cp pkg/desktop/testdata/shadcn-menu/npm/base-ui-react/LICENSE "$bundle/Contents/Resources/licenses/base-ui.txt"
cp pkg/desktop/testdata/shadcn-menu/npm/radix-ui-react-menu/LICENSE "$bundle/Contents/Resources/licenses/radix-ui.txt"
cp pkg/desktop/testdata/shadcn-menu/npm/tw-animate-css/LICENSE "$bundle/Contents/Resources/licenses/tw-animate-css.txt"
if [ "${1:-gallery}" = material-icons ]; then
    cp pkg/icons/material/sources/LICENSE "$bundle/Contents/Resources/licenses/material-symbols.txt"
fi
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
