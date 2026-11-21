// Package noreader holds cmd/adamic-performance's check that nothing that builds or tests reads Performance.jsonc. The
// line writes that file after the gate, in the landing commit itself, so anything the gate ran that read it would have
// read last landing's numbers, and a build or a test keyed on it would rerun on every landing. The check reads the
// text of every file in the repository, and it lives in a package of its own so that what it reads keys only its own
// test, as cmd/adamic/treewalk's walk of every Go file does (#hbq28nr): run from cmd/adamic-performance, it would have
// rerun that command's tests on any edit anywhere.
package noreader
