- Reran main d72728e5 with the measurement-only derived-cache fix.
- Hidden: **3,981,038 / 10,618,056 bytes (37.493097%)**.
- Delta versus 69501280: **+326,158 bytes**.
- Panic-hidden bytes: **3,647**; copier panic records: **0**.
- [Top 30, frozen-input control and complete deltas](../hidden-ranking/README.md).

6,814,108 blocked union bytes minus 2,833,070
independently examined bytes = 3,981,038 hidden bytes.
All 82 regular src/compiler files form the denominator, including comments,
whitespace and non-code inputs. [RESULT.json](RESULT.json) has every file's hashes,
byte totals, ranges and exact diagnostics. The ledger-defined union/subtraction
uses unchanged hidden.py from report commit ec0b16c0; panic boundaries are retained.

| File:lines | Hidden bytes | Stopping reasons (full diagnostics in JSON) |
|---|---:|---|
| `visitorPublic.ts:619-1798` | 60,674 | can't lower a computed field name yet |
| `parser.ts:504-1136` | 49,243 | can't lower a computed field name yet |
| `factory/nodeFactory.ts:492-1166` | 27,831 | can't lower a value of type __String yet |
| `checker.ts:51813-52186` | 26,623 | Checker-rejected body: TS2322<br>can't lower checked view field modifiers of type NodeArray<Modifier> &#124; NodeArray<ModifierLike> &#124; undefined yet |
| `program.ts:1515-1968` | 25,095 | Checker-rejected body: TS2375 |
| `checker.ts:1486-1956` | 24,935 | Checker-rejected body: TS2322 |
| `utilities.ts:11521-11907` | 24,676 | can't lower a destructured name that isn't plain yet<br>can't lower a value of type Identifier &#124; __String yet |
| `transformers/esDecorators.ts:670-1047` | 22,342 | can't lower a function returning ImmediatelyInvokedArrowFunction yet |
| `checker.ts:2046-2412` | 22,159 | Checker-rejected body: TS2322 |
| `program.ts:4073-4401` | 21,031 | Checker-rejected body: TS2345<br>Checker-rejected body: TS2375 |
