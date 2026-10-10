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
- `Inputs` is everything that can change the product: `Files` (repository-relative, hashed by content), `Flags` (every flag and environment value the build reads) and `Toolchain` (`buildcache.Tool(...)` or `runtime.Version()`). Leave one out and a stale product is served. Each cache needs a mutant that drops an input and fails a test.
- `buildcache.Product` stores the product under its key in the local cache (and Loom's shared store once the shared tier lands) and returns its directory. A product directory is always complete: a failed build leaves nothing.
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
