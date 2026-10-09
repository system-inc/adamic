# Prettier's CSS and SCSS fixtures, vendored for tests

These files are copied unchanged from Prettier's own test fixtures, `tests/format/css/**/*.css` (157 files)
and `tests/format/scss/**/*.scss` (90 files), as they stand at commit 54c8b90ae of the fork cohere's
formatter is held to (https://github.com/system-inc/prettier, forked from https://github.com/prettier/prettier).
Prettier is MIT licensed. Its LICENSE is kept beside them in LICENSE.

`css/` is the fork's `tests/format/css` and `scss/` its `tests/format/scss`, with the directory structure
kept. Only the stylesheets are here; the fork's snapshots and jsfmt specs are not, since the recorded oracle
answers in `../oracle` stand in for them. Nothing here is built into cohere's binary; it is test input for
TestPrettierFixturesMatchTheFork and TestPrettierSCSSFixturesMatchTheFork.

To refresh to another fork commit, copy the same files byte for byte from that commit (`git archive`), then
record the oracle again with `COHERE_UPDATE_ORACLE=1`.
