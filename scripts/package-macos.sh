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
    *) echo 'Expected gallery or native-popups.' >&2; exit 1 ;;
esac
mkdir -p "$bundle/Contents/MacOS"
mkdir -p "$bundle/Contents/Resources/licenses"
cp pkg/overlay/testdata/cmdk/LICENSE.md "$bundle/Contents/Resources/licenses/cmdk.txt"
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
