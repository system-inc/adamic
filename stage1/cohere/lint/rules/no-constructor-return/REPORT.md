# Fifth slot 08 batch

Ported no-constructor-return, no-delete-var and no-eq-null as .a modules in independent rule directories, with descriptors, literal messages, upstream Go adapters, witnesses and mutants. All previously claimed work was pushed through 9aec1797 before scanning 335 origin refs. The first available syntax-ready inventory rules were positions 21, 24 and 26; claim eb3b154a was pushed before implementation. The original 46 helper-ready rules were reserved. selection.json records main source checks and all-origin claim exclusions.

Observed final uncached parity: source Node, emitted JavaScript, ASan/UBSan native and upstream Go matched 240 compiler/stage1 files, 13,084,276 serialized bytes. TypeScript compiler pin 050880ce59e30b356b686bd3144efe24f875ebc8. All 76 captured upstream cases matched without exclusions: constructor 53, delete 9, null equality 14, totaling 22,587 bytes. Six constructor shape comparisons matched another 866 bytes. Owned witnesses matched 24,860 bytes. Final package selection passed in 246.522s. Overlay vet and uncached TestTheOracleCatchesOneByte passed; registry tests passed in 0.161s. Full repository gate was not run.

The initial upstream comparison failed on a string-named constructor. Go labels it Constructor while this parser labels it MethodDeclaration. The owned constructor rule now translates that shape, excludes static and generator methods, and reads accessor header tokens because this parser also labels accessor keys Constructor. Added Go comparisons pin string keys, generator keys, static string keys, getter/setter constructor keys and an object-literal string key. The failed initial log is retained as failed evidence; its timings are superseded by final parity.log. No shared parser or harness source was edited.

Mutants constructor_return_suppressed, delete_identifier_suppressed and null_inequality_only compiled and ran with exit 0 and empty stderr on source Node, emitted JavaScript and sanitized native. Only comparison of serialized output caught every mutant on each backend. Mutant gate passed in 99.90s. These suppress a real constructor return, change delete's identifier predicate and remove loose equality from null comparison, respectively.

Best of five rotating process timings, 77 compiler files plus 1,000 positive examples (78 files, 1,000 findings per rule). Native release, Node source and upstream Go, process startup included; all counts matched:

| Rule | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: |
| no-constructor-return | 580.97 | 749.49 | 4033.48 |
| no-delete-var | 599.38 | 880.15 | 4074.94 |
| no-eq-null | 601.38 | 833.03 | 3970.02 |

Setup reported Go/clang/Node/submodules ready 0s, cache warm 36s, total 36s, nproc 5. Evidence logs and validate.py hold exact gate commands and observations. All three rules have no fix or suggestion surface in Go, and the comparison verifies their unchanged applied source.

Default shared integration remains outside this unit: the checkout's registration requires rule.ts, emitted comparison and profile compilation need the scratch compatibility overlay. origin/codex/lint-harness-dot-a is available but was not merged or edited here. Reproduction uses the earlier owned compatibility.patch and rich comparison transport; shared production files remain untouched. Earlier JSX and malformed-source recovery gaps remain recorded in the original batch report. No cases from this batch are blocked.
