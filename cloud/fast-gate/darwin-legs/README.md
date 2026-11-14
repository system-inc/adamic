# Darwin legs

The gate boxes never build for darwin. A change touching a path whose first rule in `cloud/fast-gate/executors.txt` is
`darwin` (today `internal/apple/*` and `examples/apple/*`) is tested on a pool Mac, Shackler, and the gate passes its
darwin executor on that leg's record here: phase `darwin`, coverage's darwin half, and a-check on a darwin path that
stops only because the box has no darwin platform (the typed PlatformMissing stop, exit 5). Without a matching green
record each fails with "needs a darwin leg on a pool Mac (Shackler)", naming the darwinPaths digest it looked for.

The gate reads these records from its tools checkout, never from the candidate's tree and never fetched, so a change
can't vouch for itself.

## Entering a record

Records enter only through promote, Node's reviewed act; nothing else writes here. One leg is one file, `<tip>.json`,
named by the full sha of the tip the leg ran at.

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
  "darwinPaths": "<64 lowercase hex>"
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
- `darwinPaths` is 64 lowercase hex.

Anything else, a file that isn't JSON, or a field of the wrong type, is not green, and a record that isn't green never
makes the gate pass. `host`, `command`, `tools`, `counts.skipped` and `log.kept` are for the reader; the gate keeps
`host` beside the record's name in its result (`darwin_leg`).

## darwinPaths

The gate can't see a change's own tip: a candidate is rebased onto main with one parent. So it matches a record by
content. darwinPaths is the sha256 of the sorted lines

```
<mode> <object sha> <path>\n
```

one for every entry of `git ls-tree -r --full-tree <commit>` whose first rule in the tools checkout's `executors.txt`
is `darwin`. A record matches a candidate when it is green and its darwinPaths equals the candidate's. A main that moved
elsewhere still matches; a main that moved inside the darwin paths doesn't, and the leg runs again at the new content.

Compute it at the leg's tip with the gate's own code, never by hand:

```sh
python3 cloud/fast-gate/run.py --darwin-paths-digest --tree <checkout> --sha <tip> --tools <tools checkout>
```

It prints the digest and exits 0. When several records match, the gate takes the first green one in file name order.
