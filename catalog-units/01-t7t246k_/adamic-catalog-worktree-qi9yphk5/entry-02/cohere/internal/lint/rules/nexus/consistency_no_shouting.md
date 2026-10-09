# `nexus/consistency-no-shouting`

| | |
|---|---|
| **Recommendation** | **Yes**, at error |
| Findings in ahra | 37 on the morning of 2026-10-01 (32 true, 5 acronyms misread); 7 after the site sweep and the safe-list additions, all true; 0 that evening, the last one a wrapped quote now masked |
| Plugin | Nexus house rule, ported from `nexus/code-quality/lint/rules/ConsistencyNoShoutingRule.ts` and `nexus/source/text/Shouting.ts` |
| Auto-fixable | no, deliberately (the repair is a reason, which a rule cannot write) |
| Needs type information | no |

## What it checks

An uppercase token in a comment is shouting unless it is on the safe list (`allowedUppercaseTokens`
plus the `ISO` 4217 currency codes), carries a digit or an underscore, is a two-letter abbreviation,
or has no vowel. Code is masked first: fenced blocks, backticks, double-quoted strings, single-quoted
tokens, single-quoted capital phrases, `@example` blocks, and command lines.

Kirk's ruling on the 2026-10-01 review: all caps belongs in backticks, single quotes or double quotes
unless it is on the safe list. The safe list is the authority for what an acronym is.

## The dictionary detection the review proposed, measured and declined

The review's option C was to replace the vowel test with "is this token, lowercased, an English
word", so that non-word acronyms (`TUI`, `ARP`, `CAF`, `CDATA`) would stop needing names. It was measured
against the tree before being built, and it fails the parity doctrine in two ways.

1. **It silences code constants that the gate reports and the ruling wants quoted.** A dictionary
   cannot tell an acronym from a literal constant: both are non-words. Across ahra's comments, 82
   distinct vowel-bearing all-caps tokens that are not on the safe list and not English words appear
   inside backticks, which is the writers' own verdict that they are code: `SIGTERM` (43),
   `PRAGMA` (31), `SIGKILL` (24), `ENOENT`, `ENXIO`, `STDOUT`, `STDERR`, `ROWID`, `AUTOINCREMENT`,
   `NULLIF`, `EPERM`, `ESRCH`, `EAGAIN` and more. Dictionary detection would accept every one of them
   bare, so cohere would stay silent where ESLint (running the TypeScript original) reports a true
   positive. That is worse than ESLint by construction.
2. **It does not fix the case it was meant to fix.** The best word list on the machine
   (`@cspell/dict-en_us` 4.4.35, 136,630 lowercase words with inflections) contains `tui`, a New
   Zealand bird, so `TUI` would still read as shouting. `/usr/share/dict/words` (Webster's 2nd) has no
   inflected forms (`yields` is absent) and also contains `tui`.

The vocabulary gap is irreducible: the rule has to be told which non-words are nouns a sentence names
and which are literals a reader matches, and that is what the safe list is. `TUI`, `ARP`, `CAF` and `CDATA`
were added to it in both the Go port and the TypeScript original, each with its reason. `FHIR` (the
health-records interoperability standard, in www-phi-health's `ClinicalRecords.tsx`) was added to the
Go port afterwards and belongs in the original's list too.

## Deliberate divergences from the TypeScript original

- **Single-quoted capital phrases are masked.** `'RENAME COLUMN'` and `'ON CONFLICT DO NOTHING'` are
  quoted, which is the repair the ruling asks for, but the original's single-quote pattern stops at
  the first space, so it reads the words as shouting. The phrase mask accepts no lowercase inside and
  requires a word boundary outside both quotes, which is what keeps a possessive or a leading
  apostrophe (`the '90s were LOUD and the users' data`) from pairing into a mask. Zero ahra sites
  differ today, so `internal/differential/acknowledged.go` carries no entry. The same change belongs
  in `Shouting.ts` (`maskCodeAndCommands`), after the single-quoted token replace.
- **A double-quoted phrase that wraps onto the next comment line is masked.** Kirk's ruling: a quoted
  all-caps phrase is a literal. The original's double-quote pattern stops at a newline, so a block
  comment that reflowed `"the caller's job and you must NOT"` across two lines read the `NOT` as
  shouting; ahra's own `Shouting.ts:569` is that comment. The backtick mask already spans a wrap.
  `maskWrappedDoubleQuotes` pairs only the quotes the one-line pass left over, one on each of two
  adjacent lines, so a stray inch mark cannot open a span that unmasks a later quote. Nexus's
  `Shouting.ts` took the same mask after its double-quote replace (Nexus 68fc7fb, #dhaawhz), so the
  engines agree on that site.
- **Command lines are matched after the JSDoc gutter is stripped**, so a command inside a block
  comment is masked. Recorded in `maskCommandLines`.

Fixtures: `TestConsistencyNoShoutingKnowsTheAcronymsTheReviewFound` (the review's lines verbatim, both
directions), `TestConsistencyNoShoutingMasksASingleQuotedCapitalPhrase` and
`TestConsistencyNoShoutingMasksADoubleQuoteThatWraps`.
