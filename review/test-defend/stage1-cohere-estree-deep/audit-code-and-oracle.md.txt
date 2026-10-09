Starting origin/main: 12e77e8972a2e606cab6db05d84f428246a85339

CODE UNDER TEST: stage1/cohere/estree TypeScript ESTree port, entering pipeline.answer from main.run, including the source parser, Converter, Postprocessor and canonical serialization. Source Node, sanitized native and emitted JavaScript execute this port. Reached production functions were collected before mutation with NODE_V8_COVERAGE and are listed in reached-functions.json.

ORACLE: unmodified Go cohere ESTree API through testdata/oracle.go. Byte comparisons decide agreement. For orphan-decorator refusals, a handwritten diagnostic substring, failure exit and empty stdout decide the answer; that row is self-oracled. Original-library tests run Go cohere against pinned typescript-estree and Prettier, plus handwritten known-gap deltas, and never call the port. We will not mutate either external oracle.

Fixed menu: four standalone TypeScript port mutants in menu.json, written before any outcome. Empty-answer probe P1: answer returns an empty string at entry. Witness experiments W1 and W2 weaken firstDifference and refusedBeforeDeadline in Go overlays; these are the brief-authorized harness edits, not production mutants.
