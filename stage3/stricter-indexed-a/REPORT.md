Built the first five checker indexed-read witnesses on 390af985, with generated project .ts inputs.
Commits: this report and group 1 are on codex/stricter-indexed-a; exact SHAs are reported after committing.
Commands: setup succeeded in 1009.044s; group 1 go test passed in 22.672s; nproc=5.
Mutants: D054, D056, D057, D067 and D068 each erase one emitted panic and lose the named exit-70 stop.
Not covered yet in this committed group: the other 22 assigned ledger rows and the whole-program compiler build.

Group 1 proves D054, D056, D057, D067 and D068. Each source Node run has exit 0,
empty stderr, and stdout `7` when present or `undefined` when absent. Release
native, sanitized native and backend JavaScript agree with Node when present.
When absent they instead have empty stdout, exact named stderr
`adamic: panic: indexed read is absent: file:line:column` and exit 70. The IR and
actual CLI `--explain-checks` list exactly that read as checked, count
indexed-presence=1, and report trusted: 0.

Each mutant keeps the receiver and index lookup and erases only its emitted-C
panic. It must successfully compile with sanitizers. All five mutants exit 0;
D067 prints `0`, and the other four print `undefined`. The exact observation
assertion catches them. No build-warning or sanitizer failure is counted as a kill.

The manifests are input descriptions, not new Adamic programs. The harness writes
minimal .ts programs and their owning tsconfig in scratch directories, preserving
the receiver kind and index form while reducing TypeScript compiler payloads to
small object, callable or numeric values. This follows the base report's generated
.ts witness pattern and keeps the project's noUncheckedIndexedAccess=false. Source
Node runs the same .ts directly; no handwritten erased-source oracle is used.
Out-of-range is the absent array case. No compiler implementation file is edited.

Commands, all output written to logs before reading:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/stricter-indexed-a-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go test ./stage3/stricter-indexed-a -run 'TestCheckerIndexedWitnesses/(D054|D056|D057|D067|D068)$' -count=1 -timeout 10m -v > /tmp/stricter-indexed-a-group1-verified.log 2>&1
```

Setup: Node ready 0.125s, Go 0.156s, clang 0.955s, markdown dependencies 1.827s,
submodules 234.251s, Go build 1008.454s, cache warm 1008.701s, done 1009.044s.
Go 1.27.1, clang 20.1.8, Node 24.19.0, nproc=5, CPU quota 4.
Environment: /workspace/adamic-tools/env.sh. Checked-in logs retain the observations.
