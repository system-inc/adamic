# The Benchmarks Game's Go and Rust entries

The Computer Language Benchmarks Game's own Go and Rust programs, vendored so `go run ./bench -game` can measure Adamic against them on one box, all four languages measured there (bench/game.go, and bench/README.md). They are never edited, never formatted and never tuned against: each file is the archive's bytes, and `TestVendoredFilesAreTheArchivesBytes` holds every one to the sha256 [programs.json](programs.json) records for it. `.gitattributes` keeps git from converting them; the LICENSE files end their lines in CRLF, as the archive has them.

## Where they came from

The game's archive of every program, `public/download/benchmarksgame-sourcecode.zip` (blob `05648fc9a0d6`, sha256 `aabcf6726cdc14f0f45b99e5daba48584f94bbb48883fd3711a1d040474d1cb4`), in its repository at https://salsa.debian.org/benchmarksgame-team/benchmarksgame, commit `40296663ed350d5fe4a6ab5e367bab61cb77c219` (March 7, 2025), the same archive and commit as [bench/holdout](../holdout/README.md). Each file keeps the archive's path and name (`fasta/fasta.go-2.go`), so the game's name for an entry is its file's: `go-2`, or `go` for `fasta.go`, which the game numbers 1.

The game's figures, its sanity column in the report, are from `public/data/data.csv` at the same commit (blob `23821120252e`): each entry's elapsed and CPU seconds on the game's machine, a four-core i5-3330, at the game's n. Its build lines are from each entry's program page and are copied into programs.json as `buildLine`.

## License

The game's revised BSD license (3-clause): [LICENSE](LICENSE) is the archive's root license, and each program's directory holds the LICENSE the archive has beside it.

## The entries and their bases

Each program has two bases per language: the game's fastest entry, mostly several threads, and its fastest entry that runs one thread, where one exists. Where the fastest needs an instruction set beyond the baseline, the fastest portable entry stands beside it. An entry runs one thread or several by the game's own CPU time over its elapsed time.

| Program | Go fastest | Go one thread | Rust fastest | Rust one thread | Rust portable |
|---|---|---|---|---|---|
| binarytrees | go-2 | go-6 | rust-5 | none | |
| fannkuchredux | go-3 | go-8 | rust-6, needs SSSE3 | none | rust-4 |
| fasta | go-2 | go (`fasta.go`) | rust-7 | rust-3 | |
| knucleotide | go-7 | go-8 | rust-7 | none | |
| mandelbrot | go-4 | go-8 | rust-4 | none | |
| nbody | go-3 | go-3 | rust-9, needs AVX | rust-9 | rust-3 |
| regexredux | go-5 | none | rust-7 | none | |
| revcomp | go-6 | go-2 | rust (`revcomp.rust`) | none | |
| spectralnorm | go-4 | go (`spectralnorm.go`) | rust-6 | none | |

Go's portable basis for fannkuchredux and nbody is its fastest, go-3, since no Go entry needs an instruction set. pidigits isn't vendored: Adamic refuses it until it has `bigint` ([pidigits.refused.md](../holdout/pidigits.refused.md)), so no side of it is timed.

Adamic and Node run the holdout ports at the game's sizes, and the everyday sources for the three the everyday suite covers, at their everyday sizes: binarytrees is [trees.ts](../trees.ts) at depth 18 (the game's is 21), nbody is [nbody.ts](../nbody.ts) at 1,000,000 steps (50,000,000), spectralnorm is [spectral_norm.ts](../spectral_norm.ts) at n = 1,500 (5,500). revcomp reads fasta's output for 25,000,000, the holdout suite's, where the game's is for 100,000,000. Go and Rust get the same n as their first argument, and the three that read fasta's output get it on standard input with 0 as their argument, as the game runs them.

## Building them as the game did

Go 1.23.1 with `GOAMD64=v2` and a bare `go build`, and rustc 1.84.1 with `-C opt-level=3 -C target-cpu=ivybridge -C codegen-units=1`. The box holds both toolchains; -game takes their paths and never guesses one.

[go.mod](go.mod) makes this directory its own module, so the repository's module never builds these programs, and pins the one module a Go entry imports: `github.com/GRbit/go-pcre` at v1.0.0, the only release with the `Regexp.Matcher` and `Matcher.Match` that regexredux.go-5.go calls. v1.0.1, the newest before the game's January 2025 build, has neither. go.sum is made on the box, as rust-deps' Cargo.lock is, and committed from there.

go-pcre binds PCRE 1 through cgo, and regexredux.rust-7's pcre2-sys binds PCRE 2. `-pcre <prefix>` names where libpcre3 and libpcre2 are unpacked with their headers (`dpkg -x`): the two regexredux builds get pkg-config at `<prefix>/usr/lib/x86_64-linux-gnu/pkgconfig`, cgo's `-I<prefix>/usr/include` and `-L` at the libraries (with cgo's own `-O2 -g` kept), and rustc `-L native=` at them, and their runs get `LD_LIBRARY_PATH` at them, rather than an rpath, so each binary is linked as the game linked it. Without `-pcre` they build and run against the system's. Every path is a flag; none is a default.

[rust-deps](rust-deps/Cargo.toml) is a library with no code of its own that depends on every crate a Rust entry names, each pinned exactly to the newest release published before 2025-02-01 that supports rustc 1.84.1, read from crates.io's API:

| Crate | Version | Published | rust-version | Named by |
|---|---|---|---|---|
| bumpalo | 3.17.0 | 2025-01-28 | 1.71.1 | binarytrees.rust-5 |
| futures | 0.1.31 | 2021-02-24 | none | knucleotide.rust-7, written for the 0.1 line; mandelbrot.rust-4's build line passes it and its source never uses it |
| hashbrown | 0.15.2 | 2024-11-25 | 1.65.0 | knucleotide.rust-7 |
| itertools | 0.14.0 | 2024-12-31 | 1.63.0 | knucleotide.rust-7 |
| libc | 0.2.169 | 2024-12-19 | 1.63 | regexredux.rust-7 |
| memchr | 2.7.4 | 2024-06-14 | 1.61 | revcomp.rust |
| num | 0.4.3 | 2024-05-08 | 1.60 | knucleotide.rust-7 |
| num_cpus | 1.16.0 | 2023-06-29 | none | fasta.rust-7; fasta.rust-3's build line passes it and its source never uses it |
| pcre2-sys | 0.2.9 | 2024-03-24 | none | regexredux.rust-7 |
| rayon | 1.10.0 | 2024-03-24 | 1.63 | binarytrees.rust-5, fannkuchredux.rust-4 and -6, mandelbrot.rust-4, regexredux.rust-7, revcomp.rust, spectralnorm.rust-6 |
| spin | 0.6.0 | 2020-10-08 | none | fasta.rust-7; 0.7.0 to 0.9.8, all before the cutoff, are yanked |
| tokio-threadpool | 0.1.18 | 2020-02-04 | none | knucleotide.rust-7, written for the 0.1 line |

Each entry is linked with `--extern` against every crate it names, those its build line passes with `--extern` and those its source names with `extern crate` that the game found through `-L /opt/src/rust-libs`, and with `-L` at the crates' directory for what they depend on. The game's build lines name two builds each of futures and of num_cpus, by their hashes, and the versions it used aren't published, so these are the pins above. rust-deps/.cargo/config.toml builds the crates for the entries' target, ivybridge, and resolves what they depend on to releases rustc 1.84.1 supports.

On the box, once:

```
cd bench/game/rust-deps && cargo build --release --target x86_64-unknown-linux-gnu    # with rustc 1.84.1's cargo
go run ./bench -game -go <Go 1.23.1's go> -rustc <rustc 1.84.1> -crates bench/game/rust-deps/target/x86_64-unknown-linux-gnu/release/deps -pcre <prefix>
```

The `--target` keeps the crates the entries link apart from those cargo builds for build scripts: without it both land in one directory, and a crate both need, such as libc, is there twice at one version with different features (J-wme5jp2), which the runner refuses to guess between. pcre2-sys links -pcre's libpcre2 only when pkg-config finds it, and otherwise builds its bundled copy without a word: build rust-deps with pkg-config on `PATH`, `PKG_CONFIG_LIBDIR` at -pcre's `usr/lib/x86_64-linux-gnu/pkgconfig` and `PKG_CONFIG_SYSROOT_DIR` at -pcre, as the runner sets them for the Go entry.

The runner finds each crate's rlib by the `.d` file cargo writes beside it, which names the crate's version, so two versions of one crate are told apart.

## Golem, November 20, 2026

The first record (#m52sper). J-1ee69rp, a Loom measure job alone on Golem: Intel Core i7-4770S, 8 logical CPUs, Linux 6.8.0-147, Node 24.19.0, clang 20.1.8, the game's go1.23.1 and rustc 1.84.1 with its flags, the crates in rust-deps/Cargo.lock and libpcre3 and libpcre2 from noble. Adamic `8dc0d123`, which is main `9233213` with this runner. Load was 3.07/1.87/1.68 before, mostly bench's own build of adamic, and 1.40/1.54/1.57 after. `go run ./bench -game ... -samples 15 -timeout 10m`, 15 interleaved samples, every runner's output Node's byte for byte on every run, every peak the runner's own through the reaper, instructions:u the median of 3 untimed runs. The run's whole log is attached to #m52sper. The calls below are the rule above, read from that log's 15 runs a runner; the run itself printed spectralnorm against Go as a win, under the best-against-best rule this replaces.

**One thread against one thread, the headline.** Each cell is the runner's fastest to slowest wall time of 15.

| program | Adamic | Go, one thread | Rust, one thread | Node |
|---|---|---|---|---|
| binarytrees, depth 18 | 1.610 to 1.704 s | go-6 2.542 to 2.636 s: **faster**, 0.63x | no entry | 1.472 to 1.546 s: slower, 1.09x |
| spectralnorm, 1,500 | 0.328 to 0.330 s | go 0.328 to 0.330 s: **within the noise**, 1.00x | no entry | 0.408 to 0.436 s: faster, 0.80x |
| nbody, 1,000,000 | 0.197 to 0.210 s | go-3 0.099 to 0.103 s: slower, 1.99x | rust-9 0.038 to 0.057 s, which needs AVX: slower, 5.21x; rust-3 0.059 to 0.067 s: slower, 3.34x | 0.947 to 0.996 s: faster, 0.21x |
| mandelbrot, 16,000 | 50.476 to 51.568 s | go-8 21.144 to 21.331 s: slower, 2.39x | no entry | 66.080 to 68.504 s: faster, 0.76x |
| fannkuchredux, 12 | 113.892 to 116.629 s | go-8 34.600 to 35.066 s: slower, 3.29x | no entry | 27.934 to 28.093 s: slower, 4.08x |
| fasta, 25,000,000 | 19.026 to 19.459 s | go 3.408 to 3.642 s: slower, 5.58x | rust-3 1.883 to 2.088 s: slower, 10.11x | 21.875 to 23.189 s: faster, 0.87x |

Against Go's one-thread entries Adamic is faster on 1 of 6 (binarytrees), within the noise on 1 (spectralnorm) and slower on 4. Against Rust's it's slower on all 3. So on day one Adamic isn't faster than Go or Rust.

**The fastest entries, several threads, never headlined.** Adamic runs one thread, and each entry runs on all 8 CPUs. Adamic's best is 1.27x (binarytrees) to 18.4x (mandelbrot) Go's fastest entry's, and 5.2x (nbody) to 68.7x (mandelbrot) Rust's.

**The instructions say where the work is.**
- fannkuchredux runs 1,152 billion instructions:u, against go-8's 314 billion and Node's 267 billion: its flip loop keeps its counters as doubles (#k8vkmns).
- fasta runs 211 billion, against Go's 10.9 billion; for-in over a record still builds and walks its key list (#a9pr809).
- regexredux runs 435 billion against Node's 25 billion, 11.8x Node's wall: global match and replace don't use the lazy automaton (#sxg8hfd).
- nbody runs 1.76 billion, against go-3's 0.74 billion and rust-9's 0.20 billion.

**Memory.** Adamic's peak was 2.0 to 2.1 MB on every number program, 196 MB on binarytrees and 790 MB on regexredux. Node's was 70 MB to 900 MB, Go's 1.6 MB to 511 MB and Rust's 1.8 MB to 130 MB.

**Gaps.**
- knucleotide doesn't build natively: a conditional expression as a statement (#z880rwy).
- revcomp doesn't build natively: `Buffer.allocUnsafe`, which Library holds.
- pidigits is refused: it needs a bigint.
- The game has no one-thread Rust entry for seven programs, and no one-thread Go entry for regexredux.

**A lead, not yet a finding:** mandelbrot spends 15.96 s of its wall in the kernel natively and 19.11 s on Node, against Go's 0.03 s (#bnq3dr5).
