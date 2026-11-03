# The transform-mode guard after the oracle's source map

oracle/node.mjs now calls `stripTypeScriptTypes(source, { mode: 'transform', sourceMap: true, sourceUrl: url })`, so Node's
stack positions name the program's own lines (#f4qxx5d). nodeRunnerMode (fixtures_test.go) and the hook in
runner_guard_test.go accept that call beside the plain transform call, still exactly one in all.

Review, as the guard asks: on all 1,637 `.a` and `.ts` programs under stage3/fixtures, internal/oracle/testdata and oracle,
the stripped output parses to identical syntax with and without the map (TypeScript 6.0.3's printer, comments removed). The
map keeps comments and adds its `sourceMappingURL` line; it changes no code.

Mutant: `mode: 'transform'` dropped from node.mjs's call, leaving `{ sourceMap: true, sourceUrl: url }`. TestFixturesRecords
and TestFixturesRunner fail with `source Node runner changed: review the transform-mode hook`. TestTransformedNodeRunnerGuard
holds the same case as "source map without transform mode", refused, beside "transformed with a source map", accepted.
