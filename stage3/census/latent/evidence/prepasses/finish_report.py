"""Render final evidence only after both complete first-error streams have paired."""
from collections import Counter
import csv
import gzip
import hashlib
import json
from pathlib import Path
import sys

folder = Path(sys.argv[1])
def read_rows(label):
    return [json.loads(line) for line in gzip.decompress((folder/(label+'.jsonl.gz')).read_bytes()).splitlines()]
def unique_findings(rows, phase=None):
    return {(f['kind'],f['where'],f['reason'],f['text']) for row in rows[1:] for f in row['findings'] if phase is None or f['phase']==phase}
base, head = read_rows('first-error-base'), read_rows('first-error-head')
assert base[0] == head[0]
assert len(base)==len(head)==82 and not head[0]['diagnostics']
assert head[0]['latent_mode']=='first-error'
for rows in [base,head]:
    assert {r['file'] for r in rows[1:]}==set(rows[0]['sources'])
base_units = {u['where']:u for row in base[1:] for u in row['units']}
head_units = {u['where']:u for row in head[1:] for u in row['units']}
assert base_units.keys()==head_units.keys()
witnesses = [dict(where=where,name=head_units[where].get('name'),before=base_units[where]['instantiated'],after=head_units[where]['instantiated']) for where in sorted(base_units) if base_units[where]['instantiated']!=head_units[where]['instantiated']]
(folder/'unit-witness-changes.json').write_text(json.dumps(witnesses,indent=2)+'\n')
scan_before,scan_after = unique_findings(base,'refusal_scan'),unique_findings(head,'refusal_scan')
scan = dict(equal=scan_before==scan_after,before=len(scan_before),after=len(scan_after),removed=sorted(scan_before-scan_after),new=sorted(scan_after-scan_before))
(folder/'scan-comparison.json').write_text(json.dumps(scan,indent=2)+'\n')
summary={}
for label,rows in [('base',base),('head',head)]:
    summary[label]=dict(files=len(rows)-1,units=dict(Counter(u['status'] for row in rows[1:] for u in row['units'])),unique_raw_findings=dict(Counter(k[0] for k in unique_findings(rows))),events=sum(len(r['findings']) for r in rows[1:]))
(folder/'observation-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
call_counts = Counter(line.split('\t',1)[0] for line in (folder/'first-error-head-prepasses.tsv').read_text().splitlines())
assert call_counts['noteOmittedOptionals']==1,call_counts
(folder/'prepass-call-counts.json').write_text(json.dumps(dict(call_counts),indent=2)+'\n')
with (folder/'resources.csv').open('w') as output:
    writer=csv.writer(output,lineterminator="\n")
    writer.writerow(['run','child_exit','wall_seconds','user_seconds','system_seconds','peak_rss_kib','peak_observed_combined_rss_kib','resource_source'])
    for name in ['apply','cost-checker','cost-entry','first-error-base','first-error-head','core-full-base','core-full-head']:
        value=json.loads((folder/(name+'.status.json')).read_text())
        assert value['child_exit']==0 and value['stopped'] is None,(name,value)
        writer.writerow([name,value['child_exit'],value['wall_seconds'],value['user_seconds'],value['system_seconds'],value['peak_rss_kib'],value.get('peak_observed_combined_rss_kib'),value.get('resource_source','child getrusage')])
costs=json.loads((folder/'costs.json').read_text())
attempts=sum(head_units[u]['status'] in ['attempted','panic'] for u in head_units)
assert attempts==4059,attempts
estimated=costs['entry']['passes']['noteOmittedOptionals']['median_ms']*attempts/1000
counts=(folder/'counts.md').read_text()
source_manifest=json.loads((folder/'source-hashes.json').read_text())
for file,want in source_manifest.items():
    assert hashlib.sha256((Path('/workspace/scratch/sygfnmc-adapted')/file).read_bytes()).hexdigest()==want,file
scan_note = ('Refusal-scan finding sets are identical.' if scan['equal'] else 'Refusal-scan finding sets differ; scan-comparison.json retains every added and removed finding.')
text='''# Census prepasses, task sygfnmc

Branch: `typescript/census-prepasses`, cut on origin/main
`0816e18511b5e1ef4383543ee81c526fc36acfaf`. The tested code checkpoint
`a2d4b5783` was pushed before the long census. No production compiler or adaptation
change is included. Both sides use the same frozen tree from that main apply,
TypeScript 6.0.3 source `050880ce59e30b356b686bd3144efe24f875ebc8`, and the
pinned checker dependency recorded in provenance.json.

Both first-error streams reached 81 source files, with zero checker diagnostics,
4,059 attempted units and 504 generic deferrals. Headers, unit identities and
attempted eligibility match. These are isolated lowering observations, with
first-error masking and generic coverage limits, not native compilation results.

'''+counts+'''
## Full core scope

Both full-mode runs load the complete tsc entry program and select core.ts alone.
Each has 59 attempted units and 211 deferred generic declarations. Their complete
records, including witness metadata, are byte-identical:
`c5de384152cc9feb6cfe6c12dc0461d36e136c76e2f5ddc25e5eaa1be0f816bb`.
Each records 12 NotYet findings, 1 Refused finding and 8 boundaries, with zero
panics. The full pair has no cleared, new, changed or newly stopped identities.

## Passes, module lists and ownership

All four templates retain collectViewedElements, noteInheritance,
noteAccessorNames and noteOmittedOptionals in production order. The exemption
table beside the AST guard documents the census's conservative nil element key,
enum validation at declaration/use, and the absence of a whole-module execution
schedule in independent units. The element collector is called, but nil-key mode
checks every element read and requires no selective discovery or retry pipeline.
Enum objects are prepared lazily; an entryless census conservatively treats const
enum objects as built. Namespace and enum reads retain their ordinary runtime
readiness checks. No whole-program readiness coverage is claimed.

The loader preserves original input roots before an entry expands its observation
ledger. Prepasses use ESMModuleOrder, the ordering implementation moduleOrder
uses; directory drivers use the stable deduplicated union of the original roots'
orders. The list helper does not move dynamic-import refusals or cyclic read-proof
effects earlier in units. Those remain at their existing query sites.

On the two measured inputs, module membership equals Files exactly: 79 files for
checker.ts and 81 for tsc. Files keeps the entry first; ESM order places its
dependencies first. The three walking fact collectors produce sets, so that order
difference does not change their answers here. Membership can change inheritance,
accessor and omission facts on other inputs with type-only or unrelated roots.
Enum validation is order-independent; namespace readiness is order-sensitive.
collectViewedElements is a no-op with the census's nil key.

Omission discovery reads contextual checker types before explicit unit slots exist
and writes only omittedOptionals. A registered snapshot captures that program-only
map; latentUseProgramFacts installs an owned clone in every unit. Statecopy's table,
type guard, capture/install and owned-map guards include its exact symbol-keyed
shape. Statement snapshots copy it too. The historical predicates driver uses a
minimal helper with the same owned-map policy. The completed entry stream records
one omission discovery call; the other three retained calls run per fresh unit.

## Per-unit costs

Eight fresh preparations were timed for each requested entry, in production pass
order. The two exempt readiness passes were timed on disposable lowerings solely
for this probe. Neither returned an error. Both checker programs have zero
diagnostics. These sampled wall times include timer/logging overhead; the nil-key
collector mostly measures that overhead. First samples, means and ranges are in
costs.json.

'''+(folder/'costs.md').read_text()+f'''
At the tsc median, repeating omission discovery adds 25.615 ms to 39.429 ms of
inheritance/accessor walks, about 1.65 times that walking subtotal. Over 4,059
fresh units, that is approximately {estimated:.1f} seconds of repeated discovery,
before map-copy costs. The implementation computes it once. This is an estimate
from sampled warm costs, not a measured whole-census speedup.

## Guard and validation

The Go AST guard follows Lower's local delegates to the one moduleOrder-to-refuse
window, extracts every lowering method call there, and checks each template in
order against explicit exemptions. It has no Go/compiler subprocess and runs in
parallel. The omission mutant rejects noteOmittedOptionals in the named template;
the fake production pass rejects all four templates. Both mutant files retain their
in-process red assertions and diagnostic logs.

The entire latent package and predicates Go package pass, including
TestLatentSpeculativePrivateIRFields and the state-copy guards. Predicates family
and continuation/body-boundary audits, the latent no-output/continuation audit,
and all 15 pairing tests pass. Both entry overlays and the standard overlay build
successfully outside gate tests. The zero-diagnostic omission probe is missed by
main and observed twice by changed first-error, full and speculative modes, including
the second independent unit. probe-result.json retains those assertions.

The earlier absent-optional report and literal/decodeJson prepass proofs at
2c62ea4e motivated the correction; their feature compiler was not merged here.
The early cold statecopy publication timeout is retained along with the later
passing whole-package run. The inefficient full-history source clone was stopped
with its own exit 143 and replaced by a successful exact pinned shallow fetch.

## Resources, recovery and limits

Every census child exited 0. The 30-minute silence stop and 14 GiB memory ceiling
did not trigger. The core pair ran alongside the baseline; the changed entry run
started near its end under a combined census VmRSS ceiling. Wall times are
shared-host resource observations. resources.csv retains child wall/CPU/peak data.

A transport reconnect ended the baseline's original outer runner while its census
continued unchanged. adopt.py recovered its actual Linux kernel exit status from
/proc after zombie state. Its lifetime CPU ticks and sampled lifetime VmHWM are
identified explicitly; the final polling interval can miss a peak. Later runners
use detached supervisors and wait/reap their own children. Runner checks preserve
exit 7 and verify signal and memory stops report the child's -15 and reap it.

{scan_note} {len(witnesses)} generic witness flags changed;
unit-witness-changes.json retains them without normalization. Raw streams preserve
all event contexts. Raw/distinct counts are unique identities, not event multiplicity.
The pair tool performs stock TypeScript outermost-access-chain normalization.

None of the 73 production first stops would move from core.ts:2599:15: production
already runs omission discovery before its refusal scan, and production code is
unchanged. This follows from unchanged code and pass order; the 73-root production
gate was not rerun. Census unit-stop changes do not change those production results.

See commands.md, provenance.json, source-hashes.json, the two paired JSON/Markdown
ledgers, compressed raw streams and timing journals, and final-packages.log.txt.
'''
followup = folder/'ordering-error-follow-up.md'
if followup.exists():
    text = text.replace('# Census prepasses, task sygfnmc\n\n', '# Census prepasses, task sygfnmc\n\n'+followup.read_text(), 1)
(folder/'report.md').write_text(text)
