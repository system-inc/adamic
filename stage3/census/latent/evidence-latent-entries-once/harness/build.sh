#!/usr/bin/env bash
# build.sh <repository tree> <destination>: the meter's build_census for the latent binary only.
set -euo pipefail
source=$1 destination=$2
mkdir -p "$destination"
source /opt/adamic-tools/env.sh
python3 "$source/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$destination/latent-overlay.log" 2>&1
gofmt -w "$destination/latent-overlay/"*.go
(cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$destination/latent-build.log" 2>&1
echo built "$destination/latent-census"
