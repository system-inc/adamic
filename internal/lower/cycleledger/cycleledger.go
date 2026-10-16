// Package cycleledger holds the import-cycle ledger: cohere's nexus/correctness-no-import-cycle-load-time-read run over
// pristine TypeScript beside lowering's own proof of which cycle reads are safe. It is a package of its own so that its
// test, the one place Adamic runs that rule, links cohere's lint, and the compiler and lowering's tests never do
// (#321063n).
package cycleledger
