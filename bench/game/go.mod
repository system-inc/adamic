// The Benchmarks Game's Go entries, vendored here unchanged (README.md). They are their own module,
// so the repository's module never builds them, and so regexredux.go-5.go finds the one module it
// imports. go-pcre is pinned at v1.0.0, the only release with the Regexp.Matcher and Matcher.Match
// that regexredux.go-5.go calls; v1.0.1 (September 2023, before the game's January 2025 build) has
// neither. It binds PCRE 1 through cgo, so the box needs libpcre's headers and pkg-config.
module github.com/system-inc/adamic/bench/game

go 1.23

require github.com/GRbit/go-pcre v1.0.0
