#!/bin/bash
# Proves staged macOS packages check Swift with no cohere checkout on the machine.
#
#     check-swift-packages.sh <checkout> <platform>:<flag> [<platform>:<flag> ...]
#     check-swift-packages.sh "$GITHUB_WORKSPACE" darwin-arm64:--no-fix darwin-x64:--lint
#
# A released cohere runs the cohere-swift shipped beside it, never a checkout, and the engine reads
# nothing from the tree it was built in: its vocabulary comes from cohere. <checkout> is that tree, or
# wherever one could be, so it is made unreadable, and a copy of each package from ./dist checks a
# sample package from outside it, the way a user's machine runs one.
#
# Not sandbox-exec: SwiftPM sandboxes manifest compilation itself, sandboxes do not nest, and the run
# would fail for a reason that has nothing to do with the package.
#
# The flag is the run: `--no-fix` runs every phase and must cover a file in types and in lint. `--lint`
# must cover one in lint. An x64 package on Apple silicon runs under Rosetta, so an engine built for the
# wrong Mac is executed, not only inspected. It can only lint there: the types phase loads the
# toolchain's libclang, and Xcode on Apple silicon ships it for arm64 only.
set -euo pipefail

checkout=$1
shift

gate=$(mktemp -d)
mkdir -p "$gate/Sample/Sources/Sample"
printf '// swift-tools-version:6.2\nimport PackageDescription\n\nlet package = Package(name: "Sample", targets: [.target(name: "Sample")])\n' \
    > "$gate/Sample/Package.swift"
printf 'public func greeting() -> String {\n    "hello"\n}\n' > "$gate/Sample/Sources/Sample/Greeting.swift"
for run in "$@"; do
    cp -R "dist/cohere-${run%%:*}" "$gate/"
done

cd "$gate"
chmod 000 "$checkout"
trap 'chmod 755 "$checkout"' EXIT
# Proved rather than assumed: a wall that lets the read through would pass this whole check while
# showing nothing.
if ls "$checkout" > /dev/null 2>&1; then
    echo "$checkout is still readable, so this check would prove nothing"
    exit 1
fi

for run in "$@"; do
    platform=${run%%:*}
    flag=${run#*:}
    # A sample of its own, so a later run cannot pass on what an earlier one wrote.
    cp -R "$gate/Sample" "$gate/Sample-$platform"
    launch=()
    if [ "$platform" = darwin-x64 ] && [ "$(uname -m)" = arm64 ]; then
        launch=(arch -x86_64)
    fi

    status=0
    # The `+` form, because bash 3.2, the one macOS ships, calls an empty array unbound under `set -u`.
    ${launch[@]+"${launch[@]}"} "$gate/cohere-$platform/bin/cohere" "$flag" --directory "$gate/Sample-$platform" \
        > "$gate/$platform.log" 2>&1 || status=$?
    cat "$gate/$platform.log"

    # Findings exit 1, and that is a run that checked, unless the front door says the run did not
    # check or cannot be trusted. Anything above 1 did not check either.
    if [ "$status" -gt 1 ] || grep -Eq "nothing was checked|engine is broken" "$gate/$platform.log"; then
        echo "$platform $flag: the Swift run did not check the sample (exit $status)"
        exit 1
    fi
    if ! grep -Eq '^lint: .* over [1-9][0-9]* files' "$gate/$platform.log"; then
        echo "$platform $flag: the lint phase covered no file"
        exit 1
    fi
    if [ "$flag" = --no-fix ] && ! grep -Eq '^types: [0-9]+ diagnostics over [1-9][0-9]* files' "$gate/$platform.log"; then
        echo "$platform $flag: the types phase covered no file"
        exit 1
    fi
    echo "$platform $flag: checked the sample with no checkout readable"
done
