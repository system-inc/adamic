# Raw-byte predicates

Each helper owns one .a file and observes the same byte domain as Go:

| Helper | API | Contract |
| --- | --- | --- |
| isDigit | number representing a Go byte | Exactly ASCII 0..9, byte codes 48..57 |
| isBracketed | readonly number[] | At least two bytes, first 91 and last 93 |
| hasAnyPrefix | byte view plus readonly prefix views | Any exact byte prefix; an empty prefix always matches |

Adapters supply integer byte values in 0..255. These are raw Go string byte
views, not JavaScript string indexes or Unicode scalar values. Invalid UTF-8,
NUL and arbitrary high bytes are preserved. Fractional/out-of-range numbers,
undefined entries and corrupt adapter values are outside the Go byte domain.
The numeric types do not claim a new static bounded-integer type.

Nil and empty Go strings have the same empty byte view. Nil and empty prefix
lists both produce false; an empty prefix within a list produces true. Bracket
matching checks endpoints only, without validating content, nesting or balance.
Input arrays remain readonly and no helper allocates or mutates caller data.

The Go overlay exports the actual private helpers. The oracle expands captured
consumer sources and actual decoded parser literal/template fields into byte
cases, then evaluates all three predicates. Actual captured entry calls are
recorded separately, with their helper names and prefix lists. The runner uses
the existing options_json.ts only for test data; all new Adamic sources are .a.
Source Node, sanitized native and emitted JavaScript compare directly to Go.

The raw sweep covers every zero-, one- and two-byte input under five prefix
configurations. Shape controls add longer bracket/prefix data and Unicode byte
sequences. No shared compiler, registration or harness file is edited: capture
uses a temporary Go overlay. See REPORT.md for observed call limits and mutants.
