# Test products: building is its own unit

A product is anything a package's tests need built before they can run: stage 0, a checker archive, a native port binary, an oracle. Building one is its own gate unit under the same 60 s budget and 90 s kill as a test unit, and the tests that use it fetch it. Setup is still counted, in the product's unit, never in a test's.

## Declaring a product

One top-level test per product, named `TestProduct_<Name>`:

```go
func TestProduct_TsgoChecker(t *testing.T) {
	t.Parallel()
	tsgoChecker(t)
}

// The one recipe: the product test and every test that needs the checker call this.
func tsgoChecker(t testing.TB) string {
	return buildcache.Product(t, buildcache.Inputs{
		Name: "tsgo-checker",
		Files: []string{"bridge/tsgo/checker", "go.mod", "go.sum"},
		Flags: []string{"-trimpath", "GOFLAGS=" + os.Getenv("GOFLAGS")},
		Toolchain: []string{runtime.Version()},
	}, func(directory string) error {
		// build into directory
	})
}
```

- The product test and the tests share one function, so the recipe and the key can never drift apart.
- `Inputs` names the product: `Flags` (every flag and environment value the build reads) and `Toolchain` (`buildcache.Tool(...)` or `runtime.Version()`) with the name make its name key. `Files` (repository-relative) are the declared inputs: a tripwire held against what the build reads, and the key of a build for this machine alone.
- `buildcache.Product` returns the product's directory. A product directory is always complete: a failed build leaves nothing.

## Keys: what a build read

A product's key is its name key and the read set its last traced build measured, valued on the tree at hand (#vt46geg), so a change misses only the products that read it. Only Workshop's tree builder builds products for others; it opts in with `ADAMIC_BUILD_STORE=traced`, runs as main's (`ADAMIC_BUILD_STORE_TRUST=main`, without which everything it publishes lands in the candidate namespaces, `build-candidate` and `reads-candidate`, that a trusted reader never reads), and runs under the tracer:

```
ADAMIC_BUILD_STORE=traced ADAMIC_BUILD_STORE_TRUST=main go run ./internal/buildcache/cmd/traced -- go test -count=1 -run '^TestProduct_' ./...
```

`traced` runs the command under `strace -f` (Linux). Each build is given every path its test process and the processes it started opened, listed, stat'ed or missed, up to the build's end, outside keying; what the build's own processes wrote first is output, not input, and a product is an input whoever wrote it. The read set holds tree paths (a file by git's blob id, a directory by its tracked listing, a lookup by whether it exists, a symbolic link by where it points), the recipe (every file of the packages the test binary links, and its build settings), other products by their keys, the Go release and each tool by its report, and what it read of the machine (a file a dpkg package installed, unchanged since by dpkg's md5, by the package's version; anything else under /usr, /lib, /bin and /etc by itself). A machine file outside dpkg (/etc/alternatives, /usr/local) is valued when the build settles, and the run's before-and-after state covers dpkg's packages but not such files, so one changed during a traced run is keyed as it is at the end. It is recorded beside the product as `<name key>.reads`; a set recorded on the same tree as an earlier one is their union.

The store carries the read sets beside the products, as `refs/reads/<name key>.<day>.<n>`, each naming a blob that holds the whole file and the day. R2 expires refs and blobs seven days after they are written and the Worker never rewrites one it holds, so the tree builder publishes a set when it settles and again whenever it uses a product whose newest set is more than five days old; a reader takes the newest day within the lifetime, so it never depends on the oldest ref surviving. On every use the tree builder also checks that the product's ref, manifest and blobs are all still in the store, and publishes the product again, with its read sets, when any is gone. What the store still holds can't be refreshed through the Worker, which never rewrites a ref or blob it holds, so a product expires a week after it was first published and is back at the tree builder's next use; between the two a runner misses it loudly. Keeping it present throughout is Loom's to choose: a Worker that rewrites a held object's same bytes, or refs/build and blobs/ exempted from the lifecycle. A product unused by the tree builder for a week loses its read set from the store too.

A build is refused, named, and neither placed, recorded nor published, when it read what no key can name: a file git doesn't track (installed node_modules included), `.git`, content in the home directory, a temp file no process of the build made, a path outside every known place, or a lost trace. A run that wrote into the tree, or whose tree changed before it ended (`HEAD`, status, the content of every modified or untracked file, each submodule's, or the machine's installed packages), refuses every build. A refused product is Loom's to fix, never the author's red. `traced` prints each settled product with how far its declared `Files` drift from its reads, and exits 3 when the command passed and a product was refused.

Everywhere else a product is found by its read sets, the cache's and then the store's (`ADAMIC_BUILD_CACHE=read` on a runner never builds). A miss builds as it always did, for this machine alone: keyed by its declared `Files`, never found by a read set, never published.
- A product test has no subtests and asserts nothing beyond the build succeeding. It is `t.Parallel()`.

## A product's bytes name no machine

Workshop builds each product once and a runner fetches it by key into its own cache, beside its own checkout, so a product means the same wherever it is read (#tqrqx60). After every build, `buildcache` reads the product's text files and fails the build, naming the file, if one holds this machine's repository, build cache, a product directory, home, temporary directory or a directory an `ADAMIC_` variable locates. Binaries (any file with a zero byte) and macOS `.dSYM` bundles are debug information and aren't read.

- A path a product records goes through `buildcache.Relative`, which writes the tree as `<repository>`, another product (or this one, while it builds in its scratch directory) as `<build cache>/<key>`, and a gate input as `<ADAMIC_TYPESCRIPT_SOURCE>`. Its reader passes it back through `buildcache.Absolute`.
- A build input the product doesn't need afterwards, such as a `go build -overlay` JSON, leaves with the build.
- A mutant port whose imports reach back into the tree is lowered with absolute imports, then made the product's with `buildcache.RelativeFiles(directory)` as the build's last step. A reader that hands it to node or the loader takes `buildcache.Resolved(product)`, a copy beside the product whose files name this checkout.

## Using a product

A test calls the same recipe function (`tsgoChecker(t)` above). After the build phase that is a cache hit: a lookup, not a build. Start any per-test clock after the fetch. Don't add `TestX_Setup` tests, `sync.Once` setups that build, or "run Setup first; shards never build" refusals for something that can be a product. The build phase replaces them.

## What the gate does

- The fast gate, the whole gate and Loom's pool find `TestProduct_*` with `-test.list` (the fast gate also scans the source and refuses a mismatch), run each product first as its own unit on 4 CPUs under the 90 s kill, and red at `products`, naming the product, if one fails.
- Test units run with `-test.skip '^TestProduct_'` and launch only after every product has passed.
- A product over 60 s is grain like a slow test: split it into smaller products, or cache its inputs.
- Locally, `go test ./pkg` runs the product test like any other; the shared recipe means order doesn't matter.
