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

A product's key is its name key and the read set its last traced build measured, valued on the tree at hand (#vt46geg), so a change misses only the products that read it. Only Workshop builds products for others:

```
go run ./internal/buildcache/cmd/traced -- go test -count=1 -run '^TestProduct_' ./...
```

`traced` runs the command under `strace -f` (Linux). Each build is given every path its test process and the processes it started opened, listed, stat'ed or missed, up to the build's end, outside keying; what the build's own processes wrote first is output, not input. The read set holds tree paths (a file by git's blob id, a directory by its tracked listing, a lookup by whether it exists, a symbolic link by where it points), the recipe (every file of the packages the test binary links), other products by their keys, the Go release and each tool by its report. It is recorded beside the product as `<name key>.reads`, the last few sets kept; a set recorded on the same tree as an earlier one is their union.

A build is refused, named, and neither placed, recorded nor published, when it read what no key can name: a file git doesn't track (installed node_modules included), `.git`, content in the home directory, a temp file no process of the build made, a path outside every known place, a write into the tree, or a lost trace. A refused product is Loom's to fix, never the author's red. `traced` prints each settled product with how far its declared `Files` drift from its reads, and exits 3 when the command passed and a product was refused.

Anywhere else a product is found by its read sets (`ADAMIC_BUILD_CACHE=read` on a runner never builds); a miss is refused unless `ADAMIC_BUILD_STORE=off` builds it for this machine alone, keyed by its declared `Files` and never found by a read set or published.
- A product test has no subtests and asserts nothing beyond the build succeeding. It is `t.Parallel()`.

## Using a product

A test calls the same recipe function (`tsgoChecker(t)` above). After the build phase that is a cache hit: a lookup, not a build. Start any per-test clock after the fetch. Don't add `TestX_Setup` tests, `sync.Once` setups that build, or "run Setup first; shards never build" refusals for something that can be a product. The build phase replaces them.

## What the gate does

- The fast gate, the whole gate and Loom's pool find `TestProduct_*` with `-test.list` (the fast gate also scans the source and refuses a mismatch), run each product first as its own unit on 4 CPUs under the 90 s kill, and red at `products`, naming the product, if one fails.
- Test units run with `-test.skip '^TestProduct_'` and launch only after every product has passed.
- A product over 60 s is grain like a slow test: split it into smaller products, or cache its inputs.
- Locally, `go test ./pkg` runs the product test like any other; the shared recipe means order doesn't matter.
