#!/bin/sh
set -eu
if [ "$(uname -s)" != Darwin ]; then
    echo 'macOS app packaging requires macOS.' >&2
    exit 1
fi
cd "$(dirname "$0")/.."
bundle='build/Shadcn Gallery.app'
mkdir -p "$bundle/Contents/MacOS"
CGO_ENABLED=0 go build -o "$bundle/Contents/MacOS/gallery" ./cmd/gallery
cat > "$bundle/Contents/Info.plist" <<'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
    <key>CFBundleExecutable</key><string>gallery</string>
    <key>CFBundleIdentifier</key><string>dev.pylogmon.shadcn-gallery</string>
    <key>CFBundleName</key><string>Shadcn Gallery</string>
    <key>CFBundlePackageType</key><string>APPL</string>
    <key>CFBundleShortVersionString</key><string>0.1.0</string>
    <key>NSHighResolutionCapable</key><true/>
</dict></plist>
PLIST
