import json,pathlib
P=pathlib.Path(__file__).resolve().parent
J=lambda f:json.loads((P/f).read_text())
h=J('headline.json'); prov=J('provenance.json'); pre=J('preflight.json'); c=J('c-resource.json')
a=J('../a-resource.json'); b=J('../b-resource.json'); t=h['timing_totals']; pair=h['pairs']; noise=h['timing_noise']['A-C']
callers=J('caller-outcomes.json'); null=next(r for r in callers if r['name']=='nullParenthesizerRules')
def label(o):
    if o['outcome']=='stopped':return o['kind']+': '+o['reason']+' @ `'+o['where']+'`'
    return o['outcome']
def cell(v):return str(v).replace('&','&amp;').replace('<','&lt;').replace('>','&gt;').replace('|','&#124;').replace('\n',' ')
def table(headers,rows):return '\n'.join(['| '+' | '.join(headers)+' |','|'+'|'.join(['---']*len(headers))+'|']+['| '+' | '.join(cell(v) for v in row)+' |' for row in rows])+'\n'
ac=pair['A-C'];bc=pair['B-C']
bc_text=(f"B → C: **zero site, reason-row, unit-outcome or first-stop changes**. The complete C ledger is **byte-for-byte identical to B**, raw SHA256 `{prov['raw_stream_sha256']['c.jsonl']}`. This first-error census reaches no additional stop from the record-guard change. [Exhaustive site list](vs-5e5f57d3/site-movements.md), [unit list](vs-5e5f57d3/unit-changes.md) and [reason counts](vs-5e5f57d3/changed-reasons.md) are empty; JSON companions retain the empty additions and removals." if bc['unit_changes']==0 and bc['removed_distinct_sites_all_kinds']==0 and bc['added_distinct_sites_all_kinds']==0 and prov['ledger_byte_equal_to_B'] else f"B → C has {bc['unit_changes']} unit changes, {bc['removed_distinct_sites_all_kinds']} removed sites and {bc['added_distinct_sites_all_kinds']} added sites. [Every site](vs-5e5f57d3/site-movements.md), [every unit](vs-5e5f57d3/unit-changes.md), [reason counts](vs-5e5f57d3/changed-reasons.md).")
lines=['# New optimizer head 33796314: same-tree latent census','',
'**C completed all 79 files**, alone on the original box and unchanged adapted tree. The original A ledger remains the control; the original B ledger supplies the incremental comparison. All three pass the meter’s existing `report.py.lowering_summary` coverage and distinct-count validator; headers agree exactly. C exited '+str(c['exit_code'])+'.','',
table(['Distinct sites','A (head 1)','B (5e5f57d3)','C (33796314)'],[(kind,h['totals']['A'][kind],h['totals']['B'][kind],h['totals']['C'][kind]) for kind in ['NotYet','Refused','error','panic','SkippedDependency']]),
f"A-lowers → C-stops: **{ac['lowers_to_stops_other']} outside the record guard**, **{ac['lowers_to_stops_record_guard']} at the record guard**. All successful-A units absent or otherwise non-lowering on C: **{ac['lowers_to_absent_or_nonlowering']}**. [Separate record-guard exceptions with their sites](vs-a/record-guard-regressions.md), [other regressions](vs-a/other-regressions.md), [full absent/non-lowering list](vs-a/lowers-to-absent-or-nonlowering.json).",'',
f"C has **{h['C_record_guard_first_stops']} record-guard first stops**, across **{h['C_record_guard_distinct_sites']} distinct guard sites**. These seven stops are unchanged from both A and B; none is a new A-lowers/C-stops exception. The exact reason is **a union or optional field read in a program with record storage**. [Every guard first stop with A/B outcomes](record-guard-first-stops.md) and [every guard site and owner](record-guard-sites.md) include cases that were already stopped on A or B.",'',
'## First stops and sites','',
f"A → C unit first-stop/outcome changes: **{ac['unit_changes']}**; changed reason rows: **{ac['changed_reasons']}**. [Every named unit that moves](vs-a/unit-changes.md), [every changed reason with up to five locations on each side](vs-a/changed-reasons.md), and [every A lowering-phase site absent from C](vs-a/lowering-sites-absent-c.md). Reasons mentioning a view, branded supertype, intersection or predicate are marked `*` by their full wording in the reason table.",'',
bc_text,'',
'`nullParenthesizerRules`, `factory/parenthesizerRules.ts:675:1`:','',
'- A: '+label(null['A'])+'.','- B: '+label(null['B'])+'.','- C: '+label(null['C'])+'.','',
'The only changed reason row against A is the checked-predicate row quoted above, **1 → 0** at `core.ts:1799:32`. The template-interpolation row remains **2 → 2 distinct sites**, because the reached stop already appears in A through another caller. No view, branded-supertype, intersection or write-through-view reason row changes.','',
'All **78 direct cast/tryCast calls in 22 units** use the original source inventory, whose source hashes still match. [Every caller’s A/B/C outcome](caller-outcomes.md) names the unit and first stop; checker-skipped bodies are identified explicitly, with the first checker diagnostic rather than a claimed lowering result. [Full source calls and outcomes](caller-outcomes.json). Deferred-generic witness metadata changes are saved separately from first-stop movements in each pair’s `unit-metadata-changes.json`.','',
'## Time and resources','',
'**Timing design differs:** A and B were the original concurrent pair; C ran alone, as requested. These are observed per-unit differences on the same box, not an isolated cost estimate for the record guard or intersection rule. Unit time includes registration, lowering and panic recovery; whole-run time also includes loading, refusal scanning and output. All attempted units have exactly one timing row, and the timing unit populations match.','',
table(['Measurement','A','B','C'],[
('Timed units',t['A']['units'],t['B']['units'],t['C']['units']),
('Summed unit wall seconds',*[f"{t[s]['wall_seconds']:.6f}" for s in 'ABC']),
('Summed unit CPU seconds',*[f"{t[s]['cpu_seconds']:.6f}" for s in 'ABC']),
('Whole-run wall seconds',*[f"{r['wall_seconds']:.6f}" for r in [a,b,c]]),
('Peak RSS (KiB)',a['peak_rss_kib'],b['peak_rss_kib'],c['peak_rss_kib'])]),
f"C−A summed unit wall time: **{t['C']['wall_seconds']-t['A']['wall_seconds']:+.6f}s**, ratio **{t['C']['wall_seconds']/t['A']['wall_seconds']:.6f}**. [Requested 20 largest C−A differences, seconds and ratios](vs-a/timing-top20.md); [all units TSV](vs-a/unit-times.tsv); [all units including CPU JSON](vs-a/unit-times.json). The [C−B top 20](vs-5e5f57d3/timing-top20.md) and full table are also included.",'',
f"The 91 unaffected import-declaration controls have median absolute C−A difference **{noise['median_absolute_seconds']:.6f}s**, p95 **{noise['p95_absolute_seconds']:.6f}s**, maximum **{noise['max_absolute_seconds']:.6f}s**. Median C/A **{noise['median_ratio']:.3f}**; 5th–95th percentile **{noise['p05_ratio']:.3f}–{noise['p95_ratio']:.3f}**. They execute no lowered receiver read, destructure or predicate test, but still include registration cost. [Every control](vs-a/timing-controls.md), [A/C and B/C control summaries](timing-noise.json). The original simultaneous pair’s [noise floor](../timing-noise.json) remains available; shifts in these unaffected controls must be considered when reading C’s timing ratios.",'',
'Box: same kernel/hostname as the original run, nproc '+pre['nproc']+', cgroup CPU quota `'+pre['cpu_max']+'` (four cores), memory limit '+str(int(pre['memory_max'])//(1024**3))+' GiB. [Preflight](preflight.json) records the same-box and all-79-file hash checks. C had the same 5,400-second timeout and 60-second kill grace, and did not reach it. [C wall time, peak RSS, exact command and UTC bounds](c-resource.json).','',
'## Reproduction','',
'- A: `'+prov['A']+'` (original ledger and times reused byte-for-byte).',
'- B: `'+prov['B']+'` (original ledger and times reused byte-for-byte).',
'- C: `'+prov['C']+'`.',
'- Intermediate test-only commit: `148e6df3563ca142e6bda215f12ce5ae3265708b`.',
'- TypeScript: `'+prov['TypeScript']+'`; A’s original adapted tree is reused, without rebuilding.',
'- A binary SHA256: `'+prov['binary_sha256']['a']+'`.',
'- B binary SHA256: `'+prov['binary_sha256']['b']+'`.',
'- C binary SHA256: `'+prov['binary_sha256']['c']+'`.','',
'Exact commands: [preparation](preparation.sh), [prescribed overlay/build](build.sh), [alone run](run.sh), [both census-delta comparisons and validator](compare.sh). [Compiler history](compiler-history.txt) and [B→C compiler diff](compiler-diff.patch) document the tested revisions. A’s and C’s stage3 trees are identical. The worktree shared the same pinned setup dependencies; no internal/ source was edited. Only this evidence subdirectory is added to the branch.','',
'`c.jsonl.gz` and `c-times.jsonl.gz` archive the raw C streams; A/B streams remain in the parent directory. All gzip streams are CRC/roundtrip checked and their raw hashes recorded in [provenance.json](provenance.json); `SHA256SUMS` covers the new artifacts. To rerun compare.sh after packaging, first `gzip -dk c.jsonl.gz`; analyze.py can read compressed streams directly. [Validation](validation.json), [A→C delta](vs-a/census-delta.json), [B→C delta](vs-5e5f57d3/census-delta.json), logs and all derived tables are retained.','']
(P/'REPORT.md').write_text('\n'.join(lines))
