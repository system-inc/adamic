"""Render the root-only table while preserving the original observation table."""
import json
from pathlib import Path

OUT=Path(__file__).resolve().parent
ROOTS=OUT/'roots'
s=json.loads((ROOTS/'summary.json').read_text())
original=ROOTS/'original-TABLE.md'
if not original.exists():(OUT/'TABLE.md').rename(original)
esc=lambda text:text.replace('|','&#124;').replace('\n','<br>')
lines=['# Stage 3 lowering NotYet roots','',
 f"**{s['notyet_root_sites']:,} remaining NotYet root sites**, after removing **{s['echo_only_sites']:,} proven rollback echoes** from the 10,426-site census. Of the remaining sites, {s['unattributed_read_sites']:,} are unattributed reads retained conservatively; this is not a claim that every remaining site is an independent lesson.",'',
 'Compiler base remains `44583d3283fdd8674085a7ddcce040cf2a73a94e` under the authorized non-null fallback; replay tooling is `9a1f14c5`. Production compiler sources and the adapted 81-file tsc reach are unchanged. [Original scope and base evidence](EVIDENCE.md).','',
 '## Root and echo accounting','',
 f"The annotated census contains the same {s['original_notyet_sites']:,} unique lowering NotYet signatures as the original. Symbol-linked provenance identifies {s['tagged_echo_sites']:,} sites with at least one echo observation. Exactly {s['echo_only_sites']:,} are echoes in every attempt and are removed; {s['mixed_sites']:,} have both linked and unlinked attempts and remain counted. The root count is {s['original_notyet_sites']:,} - {s['echo_only_sites']:,} = {s['notyet_root_sites']:,}.",'',
 f"Echoes also reference {s['referenced_other_root_sites']:,} non-NotYet roots ({', '.join(str(v)+' '+k for k,v in s['referenced_other_roots_by_kind'].items())}). They are listed separately below and are not added to the NotYet wall. Counts deduplicate exact `(kind, where, reason, text)` root sites, not attempts. Echo counts can overlap across roots when different attempts link a site differently.",'',
 'All 1,073 reading node and 149 reading result sites are proven echoes. The 557 reading Debug sites remain unattributed reads; replay alone does not settle their independent root.','',
 'Before rollback, the measurement overlay records the failed declaration symbol and actual root finding. Later missing reads attach `blocked_by` only on resolved symbol identity, with `blocked_symbol_declaration` identifying the failed binding. Transitive echoes point directly to the first root. No spelling-based grouping is used. A successful standalone replay can still reproduce both a failed initializer and its secondary read; reproduction alone does not promote the read to a lesson.','',
 'The poison witness carries its initializer root through ordinary and shorthand reads and a secondary declaration. Shadowed names and an unrelated parameter remain usable. Dropping tagging fails the poison `blocked_by` assertion. [Commands, outputs, mutants and two real-project replays](roots/EVIDENCE.md).','',
 'Two independent nodeFactory units confirm the exact census roots: reading node at `1237:9` is blocked by `a value of type T["kind"]` at `1213:59`, and at `1246:9` by the same reason at `1436:46`. Both read signatures and governing roots reproduce. An additional checker example still reproduces its echo but reaches a different initializer failure in standalone replay; that context difference is preserved in the evidence.','',
 '| Top root reasons | Root sites | Share of remaining NotYet roots |','| ---: | ---: | ---: |']
for n in (5,10,20):
 c=s['coverage'][str(n)];lines.append(f"| {n} | {c['sites']:,} | {c['share']:.2f}% |")
lines += ['',f"{s['root_reasons']:,} root reasons; {s['root_reasons_under_five']:,} have fewer than five root sites.",'',
 'Non-null assertions, logical assignment and comma remain compiler lessons; writable-view variance remains an adaptation. Unattributed reads remain explicit pending binding-context proof. [Annotated raw CSV](roots/raw.csv), [complete root-to-echo ledger](roots/echoes.csv), [unfiltered annotated JSONL](roots/full.jsonl.gz), [root summary](roots/summary.json).','',
 '## NotYet roots, sorted by root-site count','',
 '| Root sites | Exact root reason | Echo sites | Diagnostic files with root counts | Two root examples | Context-sensitive | Disposition | Why |',
 '| ---: | --- | ---: | --- | --- | --- | --- | --- |']
for row in s['rows']:
 lines.append('| '+' | '.join([str(row['root_sites']),esc(row['reason']),str(row['echo_sites']),'<br>'.join(f'{esc(file)}: {count}' for file,count in row['files'].items()),'<br>'.join(row['examples']),'yes' if row['context_sensitive'] else 'no',row['disposition'],row['why']])+' |')
lines += ['', '## Other governing roots referenced by NotYet echoes','',
 '| Kind | Exact root reason | Root sites | Echo sites | Two root examples |','| --- | --- | ---: | ---: | --- |']
for row in s['other_roots']:
 lines.append('| '+' | '.join([row['kind'],esc(row['reason']),str(row['root_sites']),str(row['echo_sites']),'<br>'.join(row['examples'])])+' |')
lines += ['', '## Limits and preserved observations','',
 'This counts proven declaration-rollback roots and leaves unknown provenance visible. It does not infer causes across independent units or establish that untagged reads are real standalone lessons. Checker-diagnosed bodies, children behind failed compound boundaries, module order, ownership and backend execution remain outside the census. No production compiler code changed; no oracle fixture or recorded count was added.','',
 '[The original 10,426-site table](roots/original-TABLE.md), raw.csv, full.jsonl.gz and summary.json remain unchanged historical observations; their original dispositions do not override the new root/echo attribution.']
(OUT/'TABLE.md').write_text('\n'.join(lines)+'\n')
print(f"Rendered {len(s['rows'])} NotYet root rows and {len(s['other_roots'])} other governing reasons")
