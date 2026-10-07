"""Render the committed ownership verdicts into the existing census report."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parent

def render():
    rows = json.loads((ROOT / 'own-guard-verdicts.json').read_text())
    lines = ['', '## Own-property review: all 36 user-input and 43 unknown sites', '',
        'The runtime question is a **missing own key** whose spelling names an Object.prototype member. An own toString is an ordinary hit. Provenance and ownership are independent; the original 36/43 classifications remain unchanged.', '',
        '**User input (36):** 21 reads with explicit or construction/enumeration ownership proof, 11 own-check operations, two ordinary-plain-record enumeration reads, and two reads with prototype-excluding key domains. No demonstrated user-input read-first prototype miss.', '',
        '**Unknown (43):** 11 reads with ownership proof, 17 own-check operations, six ordinary-plain-record enumeration reads, four compiler-call key domains excluding prototype names, two unguarded src[e] reads with fixtures below, and three unresolved reads. No unknown site is silently counted as guarded merely because its value is tested afterward.', '',
        'Verdicts: **a** proves ownership before the value read; **a-check** is the own-property check itself (not a value read); **a-plain** proves ownership only for the ordinary unmodified prototypes used at this compiler site, because default prototype members are non-enumerable. It is not an explicit guard and does not cover arbitrary enumerable inherited properties. **safe-domain** excludes the missing prototype-name case through compiler-call key constraints, even though an ordinary missing key may still be read. These last two qualifiers are necessary: the proposed a/b/c partition alone cannot truthfully describe every operation.', '',
        '**b** reads before rejection: native cannot use an unconditional loud dictionary lookup. The exact sites needing has_own plus get_own treatment are **utilities.ts:8154:45** and **utilities.ts:8159:28**, both src[e] in compareDataObjects. Select the reject/false branch on the tested inherited misses before fetching an own value. Blindly substituting undefined is not a general semantics proof for any-valued objects. At 8154, empty objects/arrays instead give a **c** helper/API bug, so that site has the input-dependent verdict **b/c**. No full tsc CLI/watch wrong behavior from that helper was demonstrated; the two invalid empty-array configs correctly report TS5066. A blanket b lowering for all calls would conceal that distinction.', '',
        'The three **unresolved** operations are both compareProperties reads (no compiler caller of the exported core helper found) and the exotic process.env read (stock API returns a function for a missing toString, but compiler callers use internal environment names). No real tsconfig/package/command-line input reaching those sites with the required key was established. No synthetic helper call is presented as a stock CLI reproduction.', '',
        '| Location | Origin | Read/check | Verdict | Ownership evidence or limitation | Fixture |',
        '| --- | --- | --- | --- | --- | --- |']
    for row in rows:
        evidence = row['evidence'].replace('|', '\\|')
        fixture = ', '.join('[' + f + '](' + f + ')' for f in row['fixtures']) or '-'
        lines.append(f"| {row['location']} | {row['class']} | `{row['expression']}` | {row['verdict']} | {evidence} | {fixture} |")
    lines += ['', '### Stock Node observations and input fixtures', '',
        'All four keys (constructor, toString, hasOwnProperty, __proto__) are tested separately in each of seven CLI forms: missing paths, own paths, missing custom export condition with default fallback, own export condition, own typesVersions path, unknown compiler option in tsconfig, and unknown command-line option. Two additional configs exercise the empty-path-array comparison candidate. **30** unmodified stock tsc 6.0.3 runs: **29 correct, one wrong diagnostic, zero crashes, zero silently ignored cases**. Correct missing-path cases give TS2307; unknown flags give TS5023 with exit 1, unknown config options exit 2; empty arrays give TS5066.', '',
        'The one CLI candidate is an own __proto__ paths mapping lost during config conversion, not an inherited miss at any of the 79 counted sites. Package JSON preserves that own key and resolves normally. Do not infer missing-read unsafety from this separate own-key construction bug.', '',
        'Each committed input is under [input-fixtures](input-fixtures/), with its command in manifest.json and exact stdout/stderr/exit in observations.json. Source drivers are authored as .a; the runner copies unchanged text to temporary .ts/.d.ts filenames so stock tsc accepts them. The three standalone helper fixtures retain the complete upstream function and are in status.json with Node and stage-0 results. See [UPSTREAM_CANDIDATES.md](UPSTREAM_CANDIDATES.md) for the CLI candidate and the separately scoped stock helper/API candidate.', '',
        'Reproduce with STOCK_TYPESCRIPT pointing to an installed typescript@6.0.3 package: python3 probe-inputs.py (CLI observations), node probe-guards.cjs (stock API and complete ownership-row coverage), and python3 verify.py from this repository environment (19 standalone fixtures and five stdout mutants). Save output to logs; no native inherited-miss or runtime conversion mutant is claimed while these fixtures fail stage-0 checking.']
    return '\n'.join(lines) + '\n'

if __name__ == '__main__':
    print(render(), end='')
