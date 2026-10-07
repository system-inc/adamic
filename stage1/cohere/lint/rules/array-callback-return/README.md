# Array callback return

Claimed in 1f3be00e before implementation. rule.a ports callback method/argument
selection, logical/conditional/IIFE ancestry, generator and async gates, nested
return boundaries, exact head ranges and messages, allowImplicit/checkForEach/
allowVoid, and all independent suggestion edits. Go's actual rule is the oracle.

The shared control-flow graph has not been ported. graph-boundary.a is an explicit
comparison-only boundary accepting Go CFG EndReachable values for callback heads.
It does not infer flow from findings or copy any report. The private driver parses
the original source in Adamic and receives only the Go boolean per head as data.
Without that data, a checked braced value-returning callback refuses with exit 70;
it cannot silently emit zero findings. cfg-refusal.a proves the refusal on all
three runtimes. ForEach and concise callback checks need no flow data.

This is a complete rule-policy candidate with a shared CFG integration blocker,
not a certified native control-flow implementation. Native/Node measurements
consume already supplied CFG data; Go measurements build the CFG. They are not
like-for-like whole-pipeline speed comparisons. See the sibling bare-throw report
for source coverage, every excluded fixture, mutants, commands and observations.
