#!/usr/bin/env bash
set -Eeuo pipefail
set -x
# An Apple silicon Mac and an available iOS 17 or newer simulator are required.
platform=ios-sim
# shellcheck source=cloud/apple/common.sh
source "$(dirname "$0")/common.sh"
if [[ $(uname -m) != arm64 ]]; then
    echo 'apple: this arm64 simulator build requires an Apple silicon Mac' >&2
    exit 1
fi
sdk=$(xcrun --sdk iphonesimulator --show-sdk-path)
app="$out/Dedication.app"
mkdir -p "$app"
xcrun --sdk iphonesimulator clang -target arm64-apple-ios17.0-simulator -isysroot "$sdk" \
    "${flags[@]}" -I "$runtime" "$out/main.c" "${sources[@]}" -lm -o "$app/Dedication"
cat > "$app/Info.plist" <<PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>CFBundleIdentifier</key><string>$bundle_id</string>
<key>CFBundleExecutable</key><string>Dedication</string>
<key>CFBundleName</key><string>Dedication</string>
<key>CFBundlePackageType</key><string>APPL</string>
<key>CFBundleVersion</key><string>1</string>
<key>CFBundleShortVersionString</key><string>1.0</string>
<key>MinimumOSVersion</key><string>17.0</string>
<key>UIDeviceFamily</key><array><integer>1</integer><integer>2</integer></array>
<key>CFBundleSupportedPlatforms</key><array><string>iPhoneSimulator</string></array>
</dict></plist>
PLIST
plutil -lint "$app/Info.plist"
codesign --sign - --force --deep "$app"
codesign --verify --strict --verbose=2 "$app"
xcrun simctl list devices available --json > "$out/devices.json"
selection=$(node cloud/apple/simulator.mjs select "$out/devices.json")
IFS=$'\t' read -r simulator_state simulator_udid <<< "$selection"
if [[ $simulator_state != Booted ]]; then
    xcrun simctl boot "$simulator_udid"
fi
xcrun simctl bootstatus "$simulator_udid" -b
xcrun simctl install "$simulator_udid" "$app"
# Keep simctl's console transport bytes as evidence. Its launch receipt is not app stdout.
xcrun simctl launch --console "$simulator_udid" "$bundle_id" > "$out/console.stdout" 2> "$out/console.stderr"
# Remove one whole receipt line wherever simctl writes it, preserving all other bytes.
node cloud/apple/simulator.mjs filter "$out/console.stdout" "$out/app.stdout"
cmp "$out/node.stdout" "$out/app.stdout"
echo "apple: iOS simulator stdout matches Node byte for byte ($(wc -c < "$out/app.stdout") bytes)"
