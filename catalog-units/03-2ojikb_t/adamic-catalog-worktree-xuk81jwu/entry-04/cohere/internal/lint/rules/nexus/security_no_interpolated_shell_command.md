# `nexus/security-no-interpolated-shell-command`

| | |
|---|---|
| **Recommendation** | **Yes**, at `error` once ahra's 25 findings are moved to `execFileSync`/`spawn` with an argument list. Registered `off` in ahra until then |
| Findings | **ahra 25, www-phi-health 0, www-connected-app 0, api-phi-health 0** (measured 2026-10-02) |
| Measured precision | 25 of 25 true: in every finding a value the types do not prove literal reaches a shell parser. 6 are user-visible or latent bugs, 19 are harmless today |
| Plugin | cohere-native, `nexus` |
| Auto-fixable | no |
| Needs type information | yes, for the callee and for every value's literal-ness; without a checker the rule registers nothing |

## What it checks

A shell command string built at the call (a template, a `+` concatenation, a conditional, or a `const`
holding one) and handed to `node:child_process`'s `exec`/`execSync`, or to `spawn`, `spawnSync`,
`execFile` or `execFileSync` with a `shell` option typed `true` or a non-empty string, where a value
spliced into it is not typed as a string literal, number, bigint, boolean, enum member, `null` or
`undefined`. The callee is resolved by the checker's signature, so a promisified `exec` is seen and a
local `exec` is not. Values in the body of a quoted heredoc (`<<'EOSQL'`) are data to the shell and are
skipped. The rule's doc comment carries the precise version.

## Where it came from

`modules/apple/contacts/ContactsApi.ts:79` in ahra: contact search splices `LIKE '%${word}%'` built
from the searched words into `sqlite3 "${databasePath}" "${query}"`. `O'Brien` breaks the SQL and the
catch reports no contacts; `$(...)` in a name read from an email runs (`#5na28yw`). Found by both
research passes of `#tevhg3f`, built as `#ednac8q`.

## Counts against the research

ahra has 37 `exec`/`execSync` calls cohere resolves (a categorizing probe of this rule, on the same run):
25 reported, 4 built but safe, 6 all-literal, 2 opaque. No `exec` call is promisified today, and the one
`shell: true` call (`FigmaMcpLauncher.ts:22`) has only literal arguments.

- **Cross-language pass, 31 dynamic + 6 literal + 1 `shell: true`:** exact match. 31 = 25 reported +
  4 safe + 2 opaque; the 6 literal and the Figma call are silent.
- **JS pass, 20 string-typed + 8 non-literal + 3 literal-typed + 1 `shell:`:** the 20 direct templates
  are exactly cohere's 20 direct-template findings. Of its 8 non-literal commands, cohere reports 5,
  because it follows `const` bindings and `+`: `PhiCustomerApi.ts:236`, `PlanetScaleApi.ts:380`,
  `MidjourneyApi.ts:1004` and `:1021` (a `const` template) and `ConversationsBackup.ts:78` (a
  concatenation). It leaves 3: `MacOsApi.ts:128` (a `let`), `PlanetScaleApi.ts:344` (`run(command)`, a
  parameter) and `i18n-conversion.ts:243` (its only non-literal value is SQL in a quoted heredoc). The
  3 literal-typed sites are silent: `iMessageApi.ts:527` and `PhiAnalyticsApi.ts:262` (`const` string
  literals) and `AhraOsWatchers.ts:720` (a `number`). 20 + 5 = 25.

The first draft reported 26: `i18n-conversion.ts:243` was false, since the shell expands nothing in a
quoted heredoc. The heredoc scan removed it, and `PlanetScaleApi.ts:380`, whose database and branch sit
unquoted on the opener line, still reports.

## The findings, read one by one

**Real or latent bugs (6)**: a value from a person, an agent or a command line reaches the shell.

| Site | Value | Reading |
|---|---|---|
| `modules/apple/contacts/ContactsApi.ts:79` | search words in `query` | **Bug** (`#5na28yw`): an apostrophe breaks the search, `$(...)` runs |
| `modules/apple/imessage/iMessageApi.ts:135` | `parameters.sourcePath` in `"..."` | Latent: an attachment path with `"`, `$` or a backtick breaks or runs |
| `modules/apple/macos/MacOsApi.ts:134` | `finalPath`, the caller's `customPath` | Latent, same shape |
| `modules/midjourney/MidjourneyApi.ts:1004`, `:1021` | `inputPath` from the command line | Latent, same shape |
| `modules/planetscale/PlanetScaleApi.ts:380` | `database`, `branch`, `org` unquoted on the heredoc opener | Latent: command-line arguments split and expanded by `bash` |

**True but harmless today (19)**: the shell does parse a non-literal value, but where it comes from
holds no shell syntax, or it is escaped correctly.

| Site | Value | Why harmless now |
|---|---|---|
| `iMessageApi.ts:425`, `:713`, `:789`, `:857`, `:935`, `:1039`, `:1149`, `:1207`, `:1360`; `iMessageWatch.ts:185`; `MacOsApi.ts:71` | a script in `'...'` with every `'` escaped | Correct POSIX escaping (11 sites); still a shell string |
| `ContactsApi.ts:131` | `databasePath`, `pkList` | AddressBook paths found on disk; integer keys from SQLite output |
| `ContactsApi.ts:231` | `databasePath`, `digits` | `digits` is `\D` stripped |
| `ContactsImport.ts:94` | `databasePath`, the query | Code-written queries with no `"` or `$` |
| `iMessageApi.ts:668` | `scriptPath` with a `randomUUID()` | A UUID has no shell syntax |
| `ConversationsBackup.ts:78`, `:87` | `JSON.stringify(destinationPath)` | Code-built path; note `JSON.stringify` leaves `$` and backticks live |
| `PhiCustomerApi.ts:236` | `database` on the opener line | Every caller passes `'phi'`; the SQL is in a quoted heredoc |
| `SpotifyApi.ts:220` | `authUrl` | `clientId` from the credentials file, the rest `encodeURIComponent` |

**False: 0.**

## Verification

- Fixtures both ways: the real ContactsApi search before (fires on the command, through two `const`
  bindings) and after (`execFileSync('sqlite3', [databasePath, query])`, silent); 25 firing shapes, 26
  silent shapes, the older `@types/node` module layout, and the no-checker decline.
- Mutation check, 25 mutants, each through `go test -overlay` and each killed: the module-name check,
  the module-block requirement, the `__promisify__` mapping, the shell gate, the shell option's type,
  the non-empty shell string, the built gate, the built gate through `const`, literal types as safe,
  the number flag, the type-parameter constraint, flattening through `const`, the single-literal
  splice, judging a conditional by branch, the literal-union heredoc decline, the conditional heredoc
  decline, the quoted-body skip, the heredoc closing line, `<<-` tab stripping, unquoted bodies staying
  judged, the unreadable-delimiter decline, the here-string skip, several openers in order, the
  argument list under a shell, and a delimiter holding a value.
