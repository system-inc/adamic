#!/usr/bin/env bash
set -Eeuo pipefail
set -x
# Build and check a command-line app bundle, not a Cocoa application.
platform=macos
# shellcheck source=cloud/apple/common.sh
source "$(dirname "$0")/common.sh"
sdk=$(xcrun --sdk macosx --show-sdk-path)
app="$out/Dedication.app"
mkdir -p "$app/Contents/MacOS"
for architecture in arm64 x86_64; do
    xcrun --sdk macosx clang -target "$architecture-apple-macos14" -isysroot "$sdk" \
        "${flags[@]}" -I "$runtime" "$out/main.c" "${sources[@]}" -lm \
        -o "$out/Dedication-$architecture"
done
xcrun lipo -create "$out/Dedication-arm64" "$out/Dedication-x86_64" \
    -output "$app/Contents/MacOS/Dedication"
cat > "$app/Contents/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>$bundle_id</string>
<key>CFBundleExecutable</key><string>Dedication</string>
<key>CFBundleName</key><string>Dedication</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleVersion</key><string>1</string>
<key>CFBundleShortVersionString</key><string>1.0</string>
<key>LSMinimumSystemVersion</key><string>14.0</string>
<key>LSBackgroundOnly</key><true/>
</dict></plist>
PLIST
plutil -lint "$app/Contents/Info.plist"
codesign --sign - --force --deep "$app"
codesign --verify --strict --verbose=2 "$app"
# Execute directly so stdout is observable; Finder does not provide a console.
"$app/Contents/MacOS/Dedication" > "$out/app.stdout" 2> "$out/app.stderr"
cmp "$out/node.stdout" "$out/app.stdout"
echo "apple: macOS stdout matches Node byte for byte ($(wc -c < "$out/app.stdout") bytes)"
