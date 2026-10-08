# Fixture provenance

The reference implementation is new scout code under the repository license.
Raw source fixtures retain their existing source headers. Public sample sources
are extracted unchanged from the repository/commit/path listed in
`testdata/public/sources.json`; their upstream licenses apply. TypeScript corpus
fixtures are from the pinned TypeScript repositories under their upstream
licenses. Go cohere witnesses are source strings from the tests named in
`testdata/upstream/sources.json`, under cohere's license. Hashes identify the exact
bytes used. These sources are parsed as data; their scripts are never executed.
