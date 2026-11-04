# Mutant evidence for the entry overlay keeping the entry first (stage3/meter/entry_overlay.py, #3b5kcjp)
# Tests: entry_test.EntryRouteTests.test_every_module_in_the_reach_has_a_route_from_the_entry (real binaries)
#        entry_test.EntryReportTests.test_overlay_keeps_the_entry_first (static)
# Run on Golem through ahra loom proof, job 01M6J3ZQ2TFTMKXP7K2JRH1GJ8, at 277d3aa5, cohere 06af2e39, go1.27.1 linux/amd64.
# Binaries built as twice-daily.sh's build_census builds them (census, latent overlay, entry overlay, entry-census),
# then from stage3/meter: CENSUS_BINARY=<census> ENTRY_CENSUS_BINARY=<entry-census> python3 -m unittest entry_test -v

## With the entry kept first (expected pass)
fixed tests exit=0
Ran 13 tests in 0.661s
OK

## Mutant: entry_overlay.py from before this change, LatentEntryReach in the checker's GetSourceFiles order (expected red)
mutant tests exit=1
test_overlay_keeps_the_entry_first (entry_test.EntryReportTests.test_overlay_keeps_the_entry_first) ... FAIL
test_every_module_in_the_reach_has_a_route_from_the_entry (entry_test.EntryRouteTests.test_every_module_in_the_reach_has_a_route_from_the_entry) ... FAIL
AssertionError: 'p.files = []*ast.SourceFile{entry}' not found in 'func (p *Program) LatentEntryReach() {\n    p.files = nil\n    for _, source := range p.compiler.GetSourceFiles() {\n        if !source.IsDeclarationFile && !IsLibrary(source) {\n            p.files = append(p.files, source)\n        }\n    }\n}\n'
AssertionError: 3 != 0 : [{'where': 'src/second.a:1:1', 'reason': 'lower: source module has no import-declaration route: <tree>/src/second.a', 'text': 'lower: source module has no import-declaration route: <tree>/src/second.a'}, {'where': 'src/tsc/tsc.ts:3:1', 'reason': 'lower: source module has no import-declaration route: <tree>/src/tsc/tsc.ts', 'text': 'lower: source module has no import-declaration route: <tree>/
Ran 13 tests in 0.653s
FAILED (failures=2)

The route probe's 3 panics are the meter's own message, from the run of 20261103T200643Z.GK5jxe (2,279 sites on a0ba9dd5):
"lower: source module has no import-declaration route". The two real probes that already existed pass under the mutant
too, since their entry declares nothing and its one dependency is the file routes start from; the route probe is the
one that can fail.
