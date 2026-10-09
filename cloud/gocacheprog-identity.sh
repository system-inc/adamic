#!/usr/bin/env bash
# cloud/gocacheprog-identity.sh proves the shared Go build cache cannot change what the gate tests.
# cloud/setup.sh turns cmd/adamic-gocacheprog on as GOCACHEPROG by default, so the products the tests
# consume may come out of that cache. This check builds the first three, with the exact go invocation of
# their recipe in bridge/tsgo/products_test.go (buildBridgeProduct): stage 0 (TestProduct_Stage0), the
# checker archive tsgo.a (TestProduct_TsgoArchive) and the oracle (TestProduct_Oracle). The archive
# mutants, the sanitized archive, region-stage0, linkage.test and the stage1 parser and scanner oracles
# are the next products to add. Each is built five ways, and every way must give the same bytes:
#
#   baseline  GOCACHEPROG unset, a fresh GOCACHE: plain go, the reference
#   cold      through adamic-gocacheprog over an empty store, which this build fills
#   warm      a fresh GOCACHE over the store the cold build filled; no package may be compiled
#   remote    a fresh GOCACHE and an empty local store, reading the filled store over HTTP, the way a
#             fresh gate box reads the shared store; no package may be compiled
#   poisoned  one output blob corrupted (same length, one byte flipped), first in the local store and
#             then in the HTTP store: adamic-gocacheprog must name the blob as poisoned, and the build
#             must rebuild the same bytes. A poisoned blob that yields different bytes fails the check;
#             so does one the program never named (the build used it unverified, or never read it, and
#             either way the mutant proved nothing).
#
# Both stores are scratch (the HTTP one on loopback), so the check never reads or writes the shared
# store. It runs `go vet` and `go test -race` on the program first. A warm build that compiled anything
# also fails: the identity would then hold only because the cache was never used.
#
# Disk: a product's legs hold a filled store, its HTTP copy and a Go cache of gigabytes each, more than
# a small Codex /tmp. Everything (every GOCACHE, store, output, and go's own temporary directory) lives
# under one scratch root, the candidate with the most free space among the repository's parent, $HOME,
# $TMPDIR and /tmp (or ADAMIC_IDENTITY_SCRATCH alone), and each product's caches, stores and outputs are
# deleted before the next product starts; only logs and manifests stay.
#
# Exit 0: every product identical. Exit 1: a mismatch, which is about the cache. Exit 2: the check could
# not judge (a build failed, the disk filled, the HTTP store refused a write), printed as "infra: ...";
# a full disk never reads as a cache mismatch. A mismatch anywhere outranks infra elsewhere.
#
#   bash cloud/gocacheprog-identity.sh [product...]   (default: every product; Linux gate box or Codex)
#
# ADAMIC_IDENTITY_PLANT=<product> is the comparator's own mutant: that product's warm build links with
# -ldflags=-s, and the check must fail naming it, with plant-kept=yes on its line: both sides of every
# mismatch are kept under <scratch>/<product>/kept/ for a byte-diff, and the plant proves they outlive cleanup.
#
# macOS: Apple's ar stamps the archive's __.SYMDEF member with the current time, and go then writes the
# archive's hash into go.o's build ID, so two plain builds of tsgo.a a second apart differ in 22 to 24
# bytes with no cache involved (measured October 9). Every leg there, the baseline included, runs with
# ZERO_AR_DATE=1, which makes ar write zero; the comparison itself stays byte for byte. A cgo binary
# linked twice on macOS can also differ in its LC_UUID (#cchsq45); that mismatch is still a failure,
# reported with the differing byte count and both UUIDs rather than hidden.
set -euo pipefail

products="stage0 tsgo.a oracle"
usage() { echo "usage: bash cloud/gocacheprog-identity.sh [product...]   products: $products" >&2; exit 2; }
selected=${*:-$products}
for product in $selected; do
	case " $products " in *" $product "*) ;; *) usage ;; esac
done

repository=$(cd "$(dirname "$0")/.." && pwd)
# Nothing from the caller's cache setup may reach a build: each leg names its own.
unset GOCACHEPROG ADAMIC_GOCACHE_DIR ADAMIC_GOCACHE_STORE ADAMIC_GOCACHE_WRITE ADAMIC_GOCACHE_TOKEN ADAMIC_GOCACHE_TRUST ADAMIC_GOCACHE_OFF
[ "$(uname)" = Darwin ] && export ZERO_AR_DATE=1
started=$(date +%s)

freeKilobytes() { df -Pk "$1" | awk 'NR == 2 {print $4}'; }
gigabytes() { awk -v kilobytes="$1" 'BEGIN {printf "%.1fG", kilobytes / 1048576}'; }
if [ -n "${ADAMIC_IDENTITY_SCRATCH:-}" ]; then
	candidates=("$ADAMIC_IDENTITY_SCRATCH")
else
	candidates=("$(dirname "$repository")" "$HOME" "${TMPDIR:-/tmp}" /tmp)
fi
root="" rootFree=-1 considered=""
for candidate in "${candidates[@]}"; do
	[ -d "$candidate" ] && [ -w "$candidate" ] || continue
	case "$considered " in *" $candidate="*) continue ;; esac
	free=$(freeKilobytes "$candidate")
	considered="$considered $candidate=$(gigabytes "$free")"
	# Ties keep the earlier candidate: the repository's filesystem, then $HOME, then the temporary ones.
	[ "$free" -gt "$rootFree" ] && root=$candidate rootFree=$free
done
[ -n "$root" ] || { echo "identity: infra: no writable scratch root among ${candidates[*]}" >&2; exit 2; }
run=$(mktemp -d "$root/gocacheprog-identity.XXXXXX")
mkdir -p "$run/tmp"
# go puts each build's work directory under GOTMPDIR, or else TMPDIR: on a Codex box /tmp/adamic-gate.
export TMPDIR="$run/tmp" GOTMPDIR="$run/tmp"

now() {
	if [ -n "${EPOCHREALTIME:-}" ]; then
		echo "$EPOCHREALTIME"
	else
		python3 -I -c 'import time; print(f"{time.time():.3f}")'
	fi
}
seconds() { awk -v start="$1" -v end="$2" 'BEGIN {printf "%.1f", end - start}'; }
digest() { sha256sum "$1" | cut -d' ' -f1; }

echo "identity: $(cd "$repository" && git rev-parse HEAD) $(go version) GOFLAGS=$(go env GOFLAGS) CGO_ENABLED=$(go env CGO_ENABLED) CC=$(go env CC) ZERO_AR_DATE=${ZERO_AR_DATE:-}"
echo "identity: scratch $run (free:$considered)"
df -h "$run" | sed 's/^/identity: df /'

# firstError <log>: the line of a failed build that names the failure, else its last line.
firstError() {
	grep -m 1 -E 'no space left on device|^go: |: (fatal )?error|\.go:[0-9]+:[0-9]+: |^(compile|link|asm|cgo|vet): |signal: |exit status [0-9]+' "$1" || tail -n 1 "$1"
}
# program <description> <log> <command...>: runs a step on the program itself; a full disk is infra (2), any
# other failure is the program's own (1).
programStep() {
	local description=$1 log=$2
	shift 2
	(cd "$repository" && "$@") > "$log" 2>&1 && return 0
	cat "$log" >&2
	if grep -q 'no space left on device' "$log"; then
		echo "identity: infra: $description failed: $(firstError "$log")" >&2
		exit 2
	fi
	echo "identity: $description failed: $(firstError "$log")" >&2
	exit 1
}
programStep "go vet" "$run/vet.log" go vet ./cmd/adamic-gocacheprog
programStep "go test -race" "$run/test.log" go test -race -count=1 ./cmd/adamic-gocacheprog
echo "identity: go vet and go test -race ./cmd/adamic-gocacheprog passed"
# The program is built as setup.sh builds it: without itself.
program=$run/adamic-gocacheprog
programStep "building adamic-gocacheprog" "$run/program.log" go build -o "$program" ./cmd/adamic-gocacheprog

# The HTTP store: public GETs, bearer-authenticated PUTs, a conflicting ref refused with 409, the paths
# adamic-gocacheprog speaks. Each object is one flat file named for its path, so a test can poison it.
served=$run/served
mkdir -p "$served"
printf 'identity-token\n' > "$run/token"
cat > "$run/store.py" << 'PY'
import http.server, os, re, sys, tempfile
root, token, port_file = sys.argv[1:4]
valid = re.compile(r"^/(blobs|refs/gocache|refs/gocache-candidate)/[0-9a-f]{64}$")
class Store(http.server.BaseHTTPRequestHandler):
    def file(self):
        return os.path.join(root, self.path.strip("/").replace("/", "_")) if valid.match(self.path) else None
    def reply(self, status, body=b""):
        self.send_response(status)
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def do_GET(self):
        path = self.file()
        if path is None or not os.path.exists(path):
            return self.reply(404)
        with open(path, "rb") as f:
            self.reply(200, f.read())
    def do_PUT(self):
        path = self.file()
        body = self.rfile.read(int(self.headers.get("Content-Length", "0")))
        if self.headers.get("Authorization") != "Bearer " + token:
            return self.reply(401)
        if path is None:
            return self.reply(404)
        if self.path.startswith("/refs/") and os.path.exists(path):
            with open(path, "rb") as f:
                if f.read() != body:
                    return self.reply(409)
        try:
            descriptor, temporary = tempfile.mkstemp(dir=root, prefix=".put-")
            with os.fdopen(descriptor, "wb") as f:
                f.write(body)
            os.replace(temporary, path)
        except OSError as error:
            print(f"store: write failed: {self.path}: {error}", file=sys.stderr, flush=True)
            return self.reply(507)
        self.reply(201)
    def log_message(self, *arguments):
        pass
# Up to sixteen requests at once from every go process: the default listen backlog of five resets them.
class Server(http.server.ThreadingHTTPServer):
    request_queue_size = 256
    daemon_threads = True
server = Server(("127.0.0.1", 0), Store)
with open(port_file + ".tmp", "w") as f:
    f.write(str(server.server_address[1]))
os.replace(port_file + ".tmp", port_file)
server.serve_forever()
PY
python3 -I "$run/store.py" "$served" identity-token "$run/port" > "$run/store.log" 2>&1 &
storeProcess=$!
trap 'kill "$storeProcess" 2> /dev/null || true' EXIT
for _ in $(seq 100); do
	[ -s "$run/port" ] && break
	sleep 0.1
done
[ -s "$run/port" ] || { cat "$run/store.log"; echo "identity: infra: the HTTP store did not start" >&2; exit 2; }
store=http://127.0.0.1:$(cat "$run/port")

# build <product> <output directory> [extra flag]: the product's go invocation from buildBridgeProduct, with -x
# so the trace counts compiler invocations.
build() {
	local product=$1 output=$2 extra=${3:-}
	cd "$repository"
	case $product in
		stage0) go build -x ${extra:+"$extra"} -o "$output/stage0" ./cmd/adamic ;;
		tsgo.a) go build -x ${extra:+"$extra"} -buildmode=c-archive -o "$output/tsgo.a" ./bridge/tsgo/archive ;;
		oracle) go build -x ${extra:+"$extra"} -o "$output/oracle" ./bridge/tsgo/oracle ;;
	esac
}

# removeCache <GOCACHE>: deletes a Go cache by its own layout (files two levels down), then the directory.
removeCache() {
	[ -d "$1" ] || return 0
	GOCACHE="$1" go clean -cache > /dev/null 2>&1 || true
	find "$1" -mindepth 2 -maxdepth 2 -type f -delete 2> /dev/null || true
	find "$1" -mindepth 1 -maxdepth 1 -type d -empty -delete 2> /dev/null || true
	find "$1" -maxdepth 1 -type f -delete 2> /dev/null || true
	rmdir "$1" 2> /dev/null || echo "identity: warning: could not remove $1 ($(du -sh "$1" 2> /dev/null | cut -f1) left)" >&2
}

# leg <product> <leg> [VARIABLE=value...]: one build in a fresh GOCACHE, setting legStatus, legSeconds,
# legCompiles (compiler invocations in the -x trace), legManifest (sha256 of every output file) and
# productPeak (the most the scratch root held for this product, measured before the leg's cache goes).
leg() {
	local product=$1 name=$2 directory=$run/$1/$2 begun extra=""
	shift 2
	mkdir -p "$directory/output" "$directory/gocache"
	[ "${ADAMIC_IDENTITY_PLANT:-}" = "$product" ] && [ "$name" = warm ] && extra=-ldflags=-s
	begun=$(now)
	legStatus=0
	(
		export GOCACHE="$directory/gocache"
		for assignment in "$@"; do export "${assignment?}"; done
		build "$product" "$directory/output" "$extra"
	) > "$directory/build.log" 2>&1 || legStatus=$?
	legSeconds=$(seconds "$begun" "$(now)")
	legCompiles=$(grep -cE '/compile( |$)' "$directory/build.log" || true)
	(cd "$directory/output" && find . -type f | LC_ALL=C sort | while read -r file; do echo "$(digest "$file")  $file"; done) > "$directory/manifest"
	legManifest=$directory/manifest
	used=$(du -sk "$run" | cut -f1)
	[ "$used" -gt "$productPeak" ] && productPeak=$used
	# The Go cache is gone the moment the leg is measured: a product's caches run to gigabytes each.
	removeCache "$directory/gocache"
}

primary() { awk 'NR == 1 {print $1}' "$1"; }

# differs <product> <leg> <manifest>: compares a leg's outputs with the baseline's, saying which file differs.
# Both copies of every differing file are kept under $run/<product>/kept/, where discard never reaches: a
# difference that can't be byte-diffed afterwards is evidence destroyed (#vtr4at3, October 9).
differs() {
	local product=$1 name=$2 manifest=$3 file
	cmp -s "$run/$product/baseline/manifest" "$manifest" && return 1
	echo "identity: $product $name differs from baseline:" >&2
	diff "$run/$product/baseline/manifest" "$manifest" | sed 's/^/identity:   /' >&2 || true
	awk '{print $2}' "$manifest" | while read -r file; do
		local base=$run/$product/baseline/output/$file other=$run/$product/$name/output/$file
		[ -f "$base" ] && [ -f "$other" ] && ! cmp -s "$base" "$other" || continue
		mkdir -p "$(dirname "$run/$product/kept/baseline/$file")" "$(dirname "$run/$product/kept/$name/$file")"
		cp "$base" "$run/$product/kept/baseline/$file"
		cp "$other" "$run/$product/kept/$name/$file"
		echo "identity:   kept $run/$product/kept/{baseline,$name}/${file#./}" >&2
	done
	if [ "$(uname)" = Darwin ]; then
		awk '{print $2}' "$manifest" | while read -r file; do
			local base=$run/$product/baseline/output/$file other=$run/$product/$name/output/$file
			[ -f "$base" ] && [ -f "$other" ] && ! cmp -s "$base" "$other" || continue
			echo "identity:   darwin $file: $(cmp -l "$base" "$other" | wc -l | tr -d ' ') bytes differ, LC_UUID $(dwarfdump --uuid "$base" 2> /dev/null | awk '{print $2}' | head -n 1) vs $(dwarfdump --uuid "$other" 2> /dev/null | awk '{print $2}' | head -n 1) (#cchsq45)" >&2
		done
	fi
	return 0
}

# flip <file>: corrupts one byte in the middle of a file, keeping its length, so only its hash can tell.
flip() {
	python3 -I -c 'import sys
path = sys.argv[1]
data = bytearray(open(path, "rb").read())
data[len(data) // 2] ^= 0xff
open(path, "wb").write(data)' "$1"
}

# poisoned <product> <leg> <blob hash> <local blob file or "">: judges a poisoned-blob leg, echoing the outcome.
poisoned() {
	local product=$1 name=$2 hash=$3 blob=$4 log=$run/$1/$2/build.log repaired=""
	if differs "$product" "$name" "$legManifest"; then
		echo "identity: $product $name: a poisoned blob produced different bytes" >&2
		return 1
	fi
	if ! grep -q "cache poisoning: blob $hash" "$log"; then
		echo "identity: $product $name: adamic-gocacheprog never named the poisoned blob $hash" >&2
		tail -n 20 "$log" >&2
		return 1
	fi
	# The rebuild's put rewrites a local blob; a remote one stays poisoned, since the build holds no token.
	if [ -n "$blob" ]; then
		repaired=",blob-repaired=no"
		[ "$(digest "$blob")" = "$hash" ] && repaired=",blob-repaired=yes"
	fi
	echo "$name=refused,rebuilt-identical(${legCompiles}-compiles$repaired)"
}

# largest <directory> <prefix>: the hash of the largest blob there, an output rather than a record.
largest() { find "$1" -maxdepth 1 -type f -name "$2*" -exec ls -l {} + | sort -k5 -n -r | awk 'NR == 1 {print $NF}'; }

empty() { [ -d "$1" ] && find "$1" -maxdepth 1 -type f -delete; rmdir "$1" 2> /dev/null || true; }

removeStore() {
	empty "$1/blobs"
	[ -d "$1/refs" ] || { empty "$1"; return 0; }
	find "$1/refs" -mindepth 1 -maxdepth 1 -type d | while read -r namespace; do empty "$namespace"; done
	empty "$1/refs"
	empty "$1"
}

# discard <product>: everything heavy the product left, before the next product starts; logs and manifests stay.
discard() {
	local product=$1 name
	for name in baseline cold warm poisoned-local remote poisoned-remote; do
		empty "$run/$product/$name/output"
		removeCache "$run/$product/$name/gocache"
	done
	for name in store remote-store poisoned-store; do
		removeStore "$run/$product/$name"
	done
	find "$served" -maxdepth 1 -type f -delete
}

# unjudgeable <product> <leg>: true, printing "infra: ...", when the leg cannot judge the cache: its build
# failed, the disk filled, adamic-gocacheprog could not write, or the HTTP store refused a write.
unjudgeable() {
	local product=$1 name=$2 log=$run/$1/$2/build.log line
	if [ "$legStatus" != 0 ]; then
		echo "identity: infra: $product $name build failed: $(firstError "$log")" >&2
		return 0
	fi
	line=$(grep -m 1 -E 'no space left on device|adamic-gocacheprog: (write|local record|local ref)' "$log" || true)
	[ -n "$line" ] || line=$(tail -n "+$((storeLines + 1))" "$run/store.log" | grep -m 1 'store: write failed' || true)
	[ -n "$line" ] || return 1
	echo "identity: infra: $product $name: $line" >&2
	return 0
}

mismatched="" unjudged=""
for product in $selected; do
	mkdir -p "$run/$product"
	localStore=$run/$product/store
	storeLines=$(wc -l < "$run/store.log" | tr -d ' ')
	productPeak=0 verdict=ok notes="" judged=no
	common=(GOCACHEPROG="$program" ADAMIC_GOCACHE_STORE="$store" ADAMIC_GOCACHE_WRITE="$store")
	# One pass; a leg that cannot judge the cache breaks out, leaving the product unjudged.
	while true; do
		leg "$product" baseline
		unjudgeable "$product" baseline && break
		baseline=$(primary "$legManifest") baselineSeconds=$legSeconds

		leg "$product" cold "${common[@]}" ADAMIC_GOCACHE_DIR="$localStore" ADAMIC_GOCACHE_TRUST=main ADAMIC_GOCACHE_TOKEN="$run/token"
		unjudgeable "$product" cold && break
		cold=$(primary "$legManifest") coldSeconds=$legSeconds coldCompiles=$legCompiles
		differs "$product" cold "$legManifest" && verdict=MISMATCH

		leg "$product" warm "${common[@]}" ADAMIC_GOCACHE_DIR="$localStore"
		unjudgeable "$product" warm && break
		warm=$(primary "$legManifest") warmSeconds=$legSeconds warmCompiles=$legCompiles
		differs "$product" warm "$legManifest" && verdict=MISMATCH
		[ "$warmCompiles" = 0 ] || { echo "identity: $product warm compiled $warmCompiles packages: the local store did not serve them" >&2; verdict=MISMATCH; }

		hash=$(basename "$(largest "$localStore/blobs" "")")
		flip "$localStore/blobs/$hash"
		leg "$product" poisoned-local "${common[@]}" ADAMIC_GOCACHE_DIR="$localStore"
		unjudgeable "$product" poisoned-local && break
		outcome=$(poisoned "$product" poisoned-local "$hash" "$localStore/blobs/$hash") || verdict=MISMATCH
		notes="$notes ${outcome:-poisoned-local=FAILED}"
		removeStore "$localStore"

		leg "$product" remote "${common[@]}" ADAMIC_GOCACHE_DIR="$run/$product/remote-store"
		unjudgeable "$product" remote && break
		remote=$(primary "$legManifest") remoteSeconds=$legSeconds remoteCompiles=$legCompiles
		differs "$product" remote "$legManifest" && verdict=MISMATCH
		[ "$remoteCompiles" = 0 ] || { echo "identity: $product remote compiled $remoteCompiles packages: the HTTP store did not serve them" >&2; verdict=MISMATCH; }
		removeStore "$run/$product/remote-store"

		servedBlob=$(largest "$served" blobs_)
		hash=${servedBlob##*blobs_}
		flip "$servedBlob"
		leg "$product" poisoned-remote "${common[@]}" ADAMIC_GOCACHE_DIR="$run/$product/poisoned-store"
		unjudgeable "$product" poisoned-remote && break
		outcome=$(poisoned "$product" poisoned-remote "$hash" "") || verdict=MISMATCH
		notes="$notes ${outcome:-poisoned-remote=FAILED}"
		judged=yes
		break
	done
	discard "$product"
	# The plant's own check: its warm mismatch must still be on disk, both sides, after discard.
	if [ "${ADAMIC_IDENTITY_PLANT:-}" = "$product" ] && [ "$judged" = yes ]; then
		# diff exits 1 on the difference the plant made, which pipefail would turn into the script's exit.
		lost=$({ diff "$run/$product/baseline/manifest" "$run/$product/warm/manifest" || true; } | awk '/^[<>]/ {print $2, $3}' | while read -r hash file; do
			for side in baseline warm; do
				grep -qxF "$hash  $file" "$run/$product/$side/manifest" || continue
				kept=$run/$product/kept/$side/$file
				[ -f "$kept" ] && [ "$(digest "$kept")" = "$hash" ] || echo "$side/${file#./}"
			done
		done)
		[ -d "$run/$product/kept/warm" ] || lost="$lost (no kept/warm at all)"
		if [ -n "$lost" ]; then
			echo "identity: $product plant: the mismatch's outputs did not survive cleanup: $lost" >&2
			verdict=MISMATCH
			notes="$notes plant-kept=LOST"
		else
			notes="$notes plant-kept=yes"
		fi
	fi
	if [ "$judged" = no ]; then
		echo "identity: $product UNJUDGED (infra; see $run/$product) peak=$(gigabytes "$productPeak") free=$(gigabytes "$(freeKilobytes "$run")")"
		unjudged="$unjudged $product"
		continue
	fi
	echo "identity: $product baseline=$baseline cold=$cold warm=$warm remote=$remote seconds baseline=$baselineSeconds cold=$coldSeconds warm=$warmSeconds remote=$remoteSeconds compiles cold=$coldCompiles warm=$warmCompiles remote=$remoteCompiles$notes peak=$(gigabytes "$productPeak") free=$(gigabytes "$(freeKilobytes "$run")") $verdict"
	[ "$verdict" = ok ] || mismatched="$mismatched $product"
done

echo "identity: $(($(date +%s) - started))s; logs $run"
if [ -n "$mismatched" ]; then
	echo "identity: FAILED:$mismatched${unjudged:+ (also unjudged:$unjudged)}" >&2
	exit 1
fi
if [ -n "$unjudged" ]; then
	echo "identity: infra: unjudged:$unjudged (not a cache verdict)" >&2
	exit 2
fi
echo "identity: every product is byte-identical with and without adamic-gocacheprog"
