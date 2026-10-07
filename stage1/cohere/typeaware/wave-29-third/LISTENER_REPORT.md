Built: exported numeric listenerKinds metadata for six completed rules and proposed SourceFile listeners for three claimed React rules.
Commits: follows aa42ed86 on current main e8ba3d5d; final listener commit is reported in the handoff.
Commands: setup 42s, nproc 5; nine declarations match live Go registrations in 315 bytes; completed-rule controls/corpora and sanitizers pass again.
Mutants: all nine numeric-subscription mutations compile and exit 0, caught only by registration bytes; existing rule and release mutants are rechecked.
Not covered: numeric parser nodes and per-node driver/handlers are not wired; legacy whole-file/string-kind execution remains; three full React ports still lack native SSA integration.

Current main remains `e8ba3d5d81de4d3773c723914fccd4c76248b965`. The branch
was already rebased and pushed as `aa42ed86d680cb024d4c3844133d18e8e7816d4b`.
No new claim, shared parser/harness/generator edit or push to main/area branches
is part of this continuation.

Each of the six completed rule modules now exports
`listenerKinds: readonly number[]`. The declaration describes the rule's
potential enabled subscriptions. Option activation still belongs to registration:
empty naming/global options register no Go listeners. A local adapter exports
`numericListeners()` as rule/kind metadata for all nine owned names. The React
entries are explicitly proposed source listeners, not implemented validators.
Their partial graph kernels do not imply a source-file lint entry.

| Rule | Numeric kinds |
| --- | --- |
| id-denylist | 79, 80 |
| id-match | 79, 80 |
| nexus/concurrency-no-check-then-write | 247, 248, 249 |
| no-restricted-globals | 79 |
| no-setter-return | 220, 254 |
| no-shadow-restricted-names | 170, 209, 219, 232, 261, 263, 264, 274, 275, 277 |
| react-hooks/set-state-in-effect | 307 |
| react-hooks/set-state-in-render | 307 |
| react-hooks/static-components | 307 |

These values come from the pinned Go shim's numeric SyntaxKind/ast.Kind, not a
name-to-number table or guessed TypeScript-version constants. The independent
Go oracle constructs the production rule's enabled settings and enumerates its
actual `Run` listener-map keys. A nonnil checker permits registration where
required; no callback is invoked and no checker operation is performed on that
placeholder. Nine nonempty registrations are required to avoid a vacuous pass.
The native metadata probe prints the imported declarations directly, without
fetching any AST node. Native, source Node, Go and sanitized native produce 315
identical bytes and empty stderr. Each declaration's first kind is incremented
in its own mutation, including each of the three React declarations separately.
All nine mutations compile and exit 0 with empty stderr. Only comparing their
numeric registration bytes with production Go catches them. This proves metadata
checks, not fast-driver lint coverage.

The speed migration remains partial. The current shared ParseNode has only
`readonly kind: string`; it has no numeric kind field. The current Rules APIs
and binding/span helpers also use those string kinds and index lookups. A search
of all 479 fetched origin refs found only a Tailwind helper's string listener
list, not a published numeric-node/per-node driver protocol. No shared file was
changed under the ownership instruction. The newly added declarations and
adapter do not read string kinds or refetch nodes, but the old rule execution
still does. Consequently this commit does not claim compliance of the legacy
execution paths with the new speed contract or claim a speedup.

The legacy `run()` methods scan whole files. They must NOT be invoked once per
subscribed node by the incoming driver. Migration needs a shared numeric
ParseNode contract and prebuilt per-file indexes/facts, plus explicit per-node
handlers taking the already fetched node. The declaration is ready for that
registration work; it is not a substitute for those handlers. Some rules also
inspect sibling/alias writes, so those indexes need to be shared rather than
rebuilt or rescanned per callback. Replacing the shared parser or driver here
would violate this worker's territory.

The six rule modules changed, so their actual native comparisons were rerun:

- Configured naming: 272 inputs, 209 findings, 70068 identical bytes; both
  configured mutations are caught, and full native/bridge sanitizers agree.
- Default original batch: 41 controls, 15 findings, 9881 identical bytes;
  compiler 77 roots/5241 bytes and repository 287 roots/18485 bytes both agree
  with zero findings, including sanitizers. Default denylist/match mutations,
  concurrency end-offset mutation, provenance mutation and retained-registry
  mutation all remain caught. Go test PASS 122.965s.
- Second batch: 400 controls, 252 findings, seven profile groups; all groups
  and the same two frozen corpora agree with Go, including fully instrumented
  native/bridge runs. Globals, setter and shadow mutations are caught on group
  00 with normal execution and empty stderr.

Measured whole-process observations while independent checks ran concurrently:

| Batch / corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Original / compiler | 1.900722424s | 0.296694041s | 6.41x |
| Original / repository | 0.276765118s | 0.118339159s | 2.34x |
| Second / compiler | 1.943223230s | 0.355743954s | 5.46x |
| Second / repository | 0.305504644s | 0.148666297s | 2.05x |

There is no speedup claim: the driver does not use the declarations yet. Setup
finished in 42s, tools and submodules ready 0s, cache warm 42s, nproc 5, CPU quota
400000/100000 and 17.6 GB. All test output went to log files. Complete numeric
metadata outputs, nine mutation outputs, command records and completed-rule
logs are in `validation/listeners/`.

```sh
source /workspace/adamic-tools/env.sh
python3 -u stage1/cohere/typeaware/wave-29-third/check_listeners.py /workspace/wave29-listeners-final > /tmp/wave29-listeners-final.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-configured/check.py /workspace/wave29-regex-controls > /tmp/wave29-speed-configured.log 2>&1
python3 -u stage1/cohere/typeaware/wave-29-next/check.py /workspace/wave29-speed-next > /tmp/wave29-speed-next.log 2>&1
ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-regex-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' -count=1 -timeout=30m -v > /tmp/wave29-speed-default.log 2>&1
```

No full repository gate, inherited ten-rule repeat, new raw-fact mutation repeat,
new standalone lifetime suite or full emitted-JavaScript lint comparison was run.
Those implementations did not change; the original gate's released-handle check
and retained-registry mutation did rerun. The three claimed React rules still
lack native source-to-SSA and checker/capture integration, despite the previously
tested static/control kernels. Full findings/fixes/suggestions, full-rule mutants
and lint timing for them remain unfinished. No further rules were claimed.
