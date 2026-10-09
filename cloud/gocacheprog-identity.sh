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
#             must then rebuild the same bytes or fail. A poisoned blob that yields different bytes fails
#             the check; so does one the program never named (the build used it unverified, or never
#             read it, and either way the mutant proved nothing).
#
# Both stores are scratch, in this run's directory (the HTTP one on loopback), so the check never reads
# or writes the shared store. It runs `go vet` and `go test -race` on the program first. It exits
# nonzero naming every product that differed. A warm build that compiled anything also fails: the
# identity would then hold only because the cache was never used.
#
#   bash cloud/gocacheprog-identity.sh [product...]   (default: every product; Linux gate box or Codex)
#
# ADAMIC_IDENTITY_PLANT=<product> is the comparator's own mutant: that product's warm build links with
# -ldflags=-s, and the check must fail naming it.
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
run=$(mktemp -d "${TMPDIR:-/tmp}/gocacheprog-identity.XXXXXX")
started=$(date +%s)

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
echo "identity: logs $run"

(cd "$repository" && go vet ./cmd/adamic-gocacheprog) > "$run/vet.log" 2>&1 || { cat "$run/vet.log"; echo "identity: go vet failed" >&2; exit 1; }
(cd "$repository" && go test -race -count=1 ./cmd/adamic-gocacheprog) > "$run/test.log" 2>&1 || { cat "$run/test.log"; echo "identity: go test -race failed" >&2; exit 1; }
echo "identity: go vet and go test -race ./cmd/adamic-gocacheprog passed"
# The program is built as setup.sh builds it: without itself.
program=$run/adamic-gocacheprog
(cd "$repository" && go build -o "$program" ./cmd/adamic-gocacheprog)

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
        descriptor, temporary = tempfile.mkstemp(dir=root, prefix=".put-")
        with os.fdopen(descriptor, "wb") as f:
            f.write(body)
        os.replace(temporary, path)
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
[ -s "$run/port" ] || { cat "$run/store.log"; echo "identity: the HTTP store did not start" >&2; exit 1; }
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

# leg <product> <leg> [VARIABLE=value...]: one build in a fresh GOCACHE, setting legStatus, legSeconds,
# legCompiles (compiler invocations in the -x trace) and legManifest (sha256 of every output file).
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
	# The Go cache is gone the moment the leg is measured: a product's caches run to gigabytes each.
	GOCACHE="$directory/gocache" go clean -cache 2> /dev/null || true
	rmdir "$directory/gocache" 2> /dev/null || true
}

primary() { awk 'NR == 1 {print $1}' "$1"; }

# differs <product> <leg> <manifest>: compares a leg's outputs with the baseline's, saying which file differs.
differs() {
	local product=$1 name=$2 manifest=$3 file
	cmp -s "$run/$product/baseline/manifest" "$manifest" && return 1
	echo "identity: $product $name differs from baseline:" >&2
	diff "$run/$product/baseline/manifest" "$manifest" | sed 's/^/identity:   /' >&2 || true
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
	if [ "$legStatus" = 0 ] && differs "$product" "$name" "$legManifest"; then
		echo "identity: $product $name: a poisoned blob produced different bytes" >&2
		return 1
	fi
	if ! grep -q "cache poisoning: blob $hash" "$log"; then
		echo "identity: $product $name (exit $legStatus): adamic-gocacheprog never named the poisoned blob $hash" >&2
		tail -n 20 "$log" >&2
		return 1
	fi
	if [ "$legStatus" != 0 ]; then
		echo "$name=refused,failed-loudly(exit-$legStatus)"
		return 0
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

failed=""
for product in $selected; do
	mkdir -p "$run/$product"
	localStore=$run/$product/store
	verdict=ok
	notes=""

	leg "$product" baseline
	baselineSeconds=$legSeconds
	if [ "$legStatus" != 0 ]; then
		tail -n 30 "$run/$product/baseline/build.log" >&2
		echo "identity: $product: the baseline build failed (exit $legStatus); see $run/$product/baseline/build.log" >&2
		failed="$failed $product"
		continue
	fi
	baseline=$(primary "$legManifest")
	common=(GOCACHEPROG="$program" ADAMIC_GOCACHE_STORE="$store" ADAMIC_GOCACHE_WRITE="$store")

	leg "$product" cold "${common[@]}" ADAMIC_GOCACHE_DIR="$localStore" ADAMIC_GOCACHE_TRUST=main ADAMIC_GOCACHE_TOKEN="$run/token"
	cold=$(primary "$legManifest") coldSeconds=$legSeconds coldCompiles=$legCompiles
	{ [ "$legStatus" = 0 ] && ! differs "$product" cold "$legManifest"; } || verdict=MISMATCH
	grep -q 'adamic-gocacheprog:' "$run/$product/cold/build.log" && notes="$notes cold-diagnostics(see-log)"

	leg "$product" warm "${common[@]}" ADAMIC_GOCACHE_DIR="$localStore"
	warm=$(primary "$legManifest") warmSeconds=$legSeconds warmCompiles=$legCompiles
	{ [ "$legStatus" = 0 ] && ! differs "$product" warm "$legManifest"; } || verdict=MISMATCH
	[ "$warmCompiles" = 0 ] || { echo "identity: $product warm compiled $warmCompiles packages: the local store did not serve them" >&2; verdict=MISMATCH; }

	leg "$product" remote "${common[@]}" ADAMIC_GOCACHE_DIR="$run/$product/remote-store"
	remote=$(primary "$legManifest") remoteSeconds=$legSeconds remoteCompiles=$legCompiles
	{ [ "$legStatus" = 0 ] && ! differs "$product" remote "$legManifest"; } || verdict=MISMATCH
	[ "$remoteCompiles" = 0 ] || { echo "identity: $product remote compiled $remoteCompiles packages: the HTTP store did not serve them" >&2; verdict=MISMATCH; }
	empty "$run/$product/remote-store/blobs"
	find "$run/$product/remote-store/refs" -mindepth 1 -maxdepth 1 -type d 2> /dev/null | while read -r namespace; do empty "$namespace"; done
	empty "$run/$product/remote-store/refs"
	empty "$run/$product/remote-store"

	hash=$(basename "$(largest "$localStore/blobs" "")")
	flip "$localStore/blobs/$hash"
	leg "$product" poisoned-local "${common[@]}" ADAMIC_GOCACHE_DIR="$localStore"
	outcome=$(poisoned "$product" poisoned-local "$hash" "$localStore/blobs/$hash") || verdict=MISMATCH
	notes="$notes ${outcome:-poisoned-local=FAILED}"

	servedBlob=$(largest "$served" blobs_)
	hash=${servedBlob##*blobs_}
	flip "$servedBlob"
	leg "$product" poisoned-remote "${common[@]}" ADAMIC_GOCACHE_DIR="$run/$product/poisoned-store"
	outcome=$(poisoned "$product" poisoned-remote "$hash" "") || verdict=MISMATCH
	notes="$notes ${outcome:-poisoned-remote=FAILED}"

	for directory in "$localStore" "$run/$product/poisoned-store"; do
		empty "$directory/blobs"
		find "$directory/refs" -mindepth 1 -maxdepth 1 -type d 2> /dev/null | while read -r namespace; do empty "$namespace"; done
		empty "$directory/refs"
		empty "$directory"
	done
	for name in cold warm remote poisoned-local poisoned-remote; do
		empty "$run/$product/$name/output"
	done
	find "$served" -maxdepth 1 -type f -delete

	echo "identity: $product baseline=$baseline cold=$cold warm=$warm remote=$remote seconds baseline=$baselineSeconds cold=$coldSeconds warm=$warmSeconds remote=$remoteSeconds compiles cold=$coldCompiles warm=$warmCompiles remote=$remoteCompiles$notes $verdict"
	[ "$verdict" = ok ] || failed="$failed $product"
done

echo "identity: $(($(date +%s) - started))s; logs $run"
if [ -n "$failed" ]; then
	echo "identity: FAILED:$failed" >&2
	exit 1
fi
echo "identity: every product is byte-identical with and without adamic-gocacheprog"
