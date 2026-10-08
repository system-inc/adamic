# Current: stopped on optional boolean native field read

Merged origin/area/stage1-lint `ad7bd066` into stage1-hir/wip at `c7dab335`.
The parser owner’s recovery closes the prior blocker; Go and Node whole trees and
diagnostics match on the shortest TS input, and native also parses it successfully.
File kinds now follow Go source ScriptKind for every census row.

Latest native/Node parity: **966/1,465 originals, 63/63 probes**. Current Node:
**1,013/1,465 originals, 64/64 probes**; **69/69** semantic mutants pass. Current
native compilation refuses a boolean | undefined field read in Optional terminal
dumping. See the latest EVIDENCE entry, three-line reproducer, and REPORT. The
native count is the last passing checkpoint, not current-tip certification.

Unit 2 remains unfinished: patterns, method/function variants and remaining
construction paths, rule-owned ForFunction/compilation-unit integration, and
static-components registration/certification are pending. No parser patch was
made, no failing admitted graph dropped, no native language workaround applied.
The user authorizes a single WIP push of this stopped unit.
