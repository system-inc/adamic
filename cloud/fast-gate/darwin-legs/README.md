# Darwin legs

The gate boxes never build for darwin. A change touching a path whose first rule in `cloud/fast-gate/executors.txt` is
`darwin` (today `internal/apple/*` and `examples/apple/*`) is tested on a pool Mac, Shackler, and the gate passes its
darwin executor on that leg's record here: phase `darwin` and coverage's darwin half pass on it, and a-check judges a
darwin .a that stops only because the box has no darwin platform (the typed PlatformMissing stop, exit 5) on what the
leg recorded `adamic c` saying for it on the Mac (`aCheck`, below). Without a matching green record phase `darwin` and
coverage fail with "needs a darwin leg on a pool Mac (Shackler)", naming the darwinPaths digest they looked for, and
a-check stays red as it always was. When the gate can't compute the candidate's darwinPaths (below), they fail with "no
darwin leg can match", naming why, and a-check stays red the same way; no record can pass a digest that doesn't exist.

The gate reads these records from its tools checkout, never from the candidate's tree and never fetched, so a change
can't vouch for itself.

## Entering a record

Records enter only through promote, Node's reviewed act; nothing else writes here. One leg is one file, `<tip>.json`,
named by the full sha of the tip the leg ran at. Promote checks, before a record lands:

- the tip resolves in adamic;
- its darwinPaths recomputes equal with `--darwin-paths-digest`, run in a checkout at the tip (below);
- its `aCheck` names every darwin .a in the tip's tree, the list `--darwin-a-paths` prints at the tip.

## The record

```
{
  "tip": "<40 lowercase hex, equal to the file name>",
  "host": "Shackler",
  "claim": "<claim id>",
  "log": {"sha256": "<64 lowercase hex>", "kept": "<where the log is kept>"},
  "command": "<the exact command the leg ran>",
  "tools": {"go": "...", "node": "...", "xcode": "...", "sdk": "..."},
  "counts": {"passed": <N above 0>, "failed": 0, "skipped": <N>},
  "exit": 0,
  "darwinPaths": "<64 lowercase hex>",
  "aCheck": {"<path>": {"exit": <adamic c's exit>, "errors": "<adamic c's stderr, whole up to 16384 characters>"}}
}
```

A record names exactly one of `claim` or `job` (`"job": "J-..."`). The claim form stands in for a door job only until
the macOS runner (#ez7qyry) ships; after that a leg names its job.

## Green

A record is green only when all of these hold:

- `tip` is 40 lowercase hex and equals the file name without `.json`;
- exactly one of `claim` and `job` is present, and it is a non-empty string;
- `log.sha256` is 64 lowercase hex;
- `exit` is the number 0;
- `counts.failed` is the number 0 and `counts.passed` is a number above 0;
- `darwinPaths` is 64 lowercase hex;
- `aCheck`, when present, is an object whose every value is an object holding a whole number `exit` and a string
  `errors`. A leg of a change with no darwin .a may leave `aCheck` out.

Anything else, a file that isn't JSON, or a field of the wrong type, is not green, and a record that isn't green never
makes the gate pass. `host`, `command`, `tools`, `counts.skipped` and `log.kept` are for the reader; the gate keeps
`host` beside the record's name in its result (`darwin_leg`).

## darwinPaths

The gate can't see a change's own tip: a candidate is rebased onto main with one parent. So it matches a record by
content, every input the leg's `go test ./internal/apple/...` builds from, so an equal digest means equal inputs (Adamic
Loom's ruling). darwinPaths is the sha256 of the sorted lines

```
dir <path> <object sha>\n
file <path> <blob sha>\n
<mode> <object sha> <path>\n
```

- `dir`: one for every directory inside the checkout that
  `GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 GOFLAGS=-mod=readonly go list -deps -test -f '{{.Dir}}' ./internal/apple/...`
  names, the Mac's build of the package with its tests' imports, so its whole in-module dependency closure
  (internal/lower, internal/ir, internal/native and more). Each is the path relative to the checkout, listed once, and
  `git rev-parse <commit>:<path>`, its tree object. A directory at or inside a gitlink (a submodule) is the gitlink's
  path and its commit object instead. The standard library and modules outside the checkout drop out; the module files
  pin those. A tree object takes the directory's subdirectories too, conservatively: a main that moved a listed
  package's testdata, or a package below one, moves the digest though go may read neither.
- `file`: one for each of `go.mod`, `go.sum`, `go.work` and `go.work.sum` present at the commit's root.
- `<mode>`: one for every entry of `git ls-tree -r --full-tree <commit>` whose first rule in the tools checkout's
  `executors.txt` is `darwin`.

A record matches a candidate when it is green and its darwinPaths equals the candidate's. A main that moved outside the
leg's inputs still matches; a main that moved inside them, a darwin path or any directory of the closure, doesn't, and
the leg runs again at the new content.

go list reads the files on disk, so the digest is computed only in a checkout at the commit: `git rev-parse HEAD` there
must be the full sha of the commit `--sha` names, or it's refused. The gate's tree is the candidate's checkout, so this
holds there. A go list that fails, names nothing, names no directory inside the checkout, or names one the commit's tree
doesn't hold (an untracked one) is refused too, never a digest.

Compute it in a checkout at the leg's tip with the gate's own code, never by hand:

```sh
git -C <checkout> checkout --detach <tip>
python3 cloud/fast-gate/run.py --darwin-paths-digest --tree <checkout> --sha <tip> --tools <tools checkout>
```

It prints the digest and exits 0, or prints nothing and exits 1 saying why on stderr. When several records match, the
gate takes the first green one in file name order.

## aCheck

The PlatformMissing stop comes at load, before the checker and the refusal pass run, so a gate box never learns what a
darwin .a's types are, and the leg's `go test ./internal/apple/...` never checks `examples/apple/*.a`. The Mac does it:
at the tip it builds `go build -o adamic ./cmd/adamic`, then runs `adamic c <path>` for every path

```sh
python3 cloud/fast-gate/run.py --darwin-a-paths --tree <checkout> --sha <tip> --tools <tools checkout>
```

prints (each darwin .a no `a-check-exempt` line names), recording each run's exit and stderr under its path. a-check
judges a recorded run by its own rules and the file's `// a-check:` header, exactly as a run on the box: a clean result
or a "can't lower ... yet" stop is checked, a header's refusal or type error must be met. A path the record's `aCheck`
leaves out stays red ("the darwin leg <record> recorded no adamic c for this path"), and so does a recorded exit 5,
the Mac lacking the platform too. a-check's results row names the record as `source`.
