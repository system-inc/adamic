# Mutants for an iterator step result's key order (#7nw4kgj)

Each mutant is one exact line edit, run on a Linux box against TestIteratorResultKeysAgreeWithNode
(internal/oracle/iterator_result_order_test.go) with `ADAMIC_GATE_UNCACHED=1`, then put back from a
saved copy. Each restores the old done-first order of one runtime shape's names, which is what a
program enumerating the result saw before the fix.

| Mutant | Edit | Red |
|---|---|---|
| The collection result's old order | runtime/map_set.c: `static const char *const result_names[] = {"value", "done"};` made `{"done", "value"}` | collections.a, every row: a Map's keys(), values() and entries() and a Set's values() print done,value at a step and at done. string_next shares this shape, and no program reaches its result yet: an explicit Symbol.iterator call on a string stops NotYet in lowering |
| matchAll's result's old order | runtime/regexp.c: `static const char *const next_names[] = {"value", "done"};` made `{"done", "value"}` | matches.a, both rows: matchAll's next() prints done,value at a step and at done (#7nw4kgj's follow-up pin) |
