# Lane call census

The complete original lane control uses stage3/lane/run.sh from unchanged main.
The counted tree is copied from its completed adapted tree. A source-hash-pinned,
measurement-only wrapper counts owner entries and completions. It is built with
the same lockfile and full compiler/test build. No counter is in adapt.cjs.

The single lane union is the unchanged 398-shard manifest, run in 20 pieces of at
most 20 shards with four workers and the house measurement's 90-second per-shard
deadline. Each piece is below 25 minutes. Every shard passes its ordinary test
identity, baseline and API guards. The merged identity multiset and raw baseline
diff are identical to the whole control: 106,366 passing, one sanctioned API
failure, zero skipped. Preparation/build calls and Node witnesses are excluded.

| Function | Calls | Completed | Counterexamples |
| --- | ---: | ---: | ---: |
| groupBy | 1,604 | 1,604 | 0 |
| clone | 8,603,902 | 8,603,902 | 0 |

The compiler callback isSyntheticMetadataDecorator accounts for 1,424
groupBy calls. Selector-name and caller-frame counts, result-key combinations,
input classes, per-process counts and per-piece totals are retained in
count-summary.json and the raw evidence.

Both selected adaptations emit exactly the same JavaScript as the original,
including all nine complete compiler outputs. Thus every counted input has the
same result and exception behavior. This comparison uses byte identity instead
of evaluating callbacks or getters a second time. There are zero counterexamples.
The counted source and all built files are frozen and rehashed after every piece.

Observed groupBy input classes: {"dense": 1541, "empty": 0, "sparse": 0, "undefined": 63}.
Observed clone input classes: {"nullish": 5, "object": 8603897, "primitive": 0}.
Node witnesses include undefined and empty inputs, sparse arrays, booleans,
string keys, nullable keys, prototype-key exceptions, primitives, inherited
properties and symbol-key objects as appropriate to each function. The selected
JavaScript is unchanged on the unobserved classes too.
