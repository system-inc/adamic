Built: all 21 owned rule.json manifests now use typescript-go ast.Kind names, superseding their historical numeric form.
Commits: accompanying change on codex/typeaware-wave-23, based on tested landing 9200816d25e0f3133f680fc4f797dc0bf994721e.
Commands and outputs: TestWave23NumericListeners PASS 12.614s; three JSX named registrations agree with independent Go; diff check passes.
Mutants: eighteen compiled internal listener-probe mutants caught; named JSON mutant Identifier in place of ThrowStatement failed the manifest oracle and was restored.
Not covered: JSX decision ports remain blocked on shared frontend integration; three native analysis rules remain parked; no new claims or speed improvement.

Fetched every origin head explicitly. Main and area/stage1-lint remain c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06, included in this branch. ab70f38d4 remains outside those integration branches. The fifteen rule implementations and native driver are unchanged, so the 476.138s landing comparison remains applicable to those runtime sources.

Only worker-owned manifests, independent Go registration oracles and their owned check were edited. The Go oracles add --listener-names using ast.Kind.String() with its Kind prefix removed. Existing numeric output is retained solely to compare the internal .a probes; registry manifests contain names, never numbers. These are declaration metadata, not runnable registry integrations. The existing eighteen-rule check compares named JSON declarations against production Go, then checks the internal probes in native release, sanitizer and emitted JavaScript modes, with successfully compiled mutants. The three JSX names were separately checked against their independent production Go oracle. No shared generator or harness was edited.

Commands (toolchain environment sourced):

```sh
ADAMIC_WAVE23_LISTENER_ARTIFACTS=/workspace/wave-23/named-kinds go test ./stage1/cohere/typeaware -run '^TestWave23NumericListeners$' -count=1 -timeout=10m -v > /workspace/wave-23/named-kinds.log 2>&1
# Independent JSX overlay built, then invoked with its frozen config and manifest and --listener-names.
# Intentional JSON mutation invokes the same owned test and expects failure, restoring exact original bytes.
```

Raw logs are compressed under validation-wave23-named-kinds. The named manifest mutant fails with rule.json kinds differ from Go: Identifier versus ThrowStatement. Earlier numeric reports are historical evidence, not the current manifest contract. No new rule-decision parity or full gate rerun is claimed. Native versus Go timing observations remain in WAVE_23_LANDING_C019_REPORT.md.
