"""Recount unique sites, depth tables, source coverage, and independent AST depth."""
from collections import Counter, defaultdict
import gzip
import hashlib
import html
import json
from pathlib import Path
import sys

root, observations, stock_path, output = map(Path, sys.argv[1:5])
rows = [json.loads(line) for line in observations.read_text().splitlines()]
stock = json.loads(stock_path.read_text())
assert rows[0]['latent_mode'] == 'speculative'
assert stock['typescript'] == '6.0.3'
stock_files = {r['file']: r for r in stock['files']}
counts = {kind: Counter() for kind in ('NotYet', 'Refused')}
reasons = defaultdict(Counter)
seen = set()
files = []
depth_matches = depth_unmatched = depth_without_source = 0
span_indices = {}
def stock_spans(filename):
    relative = str(Path(filename).relative_to(root))
    if relative not in stock_files:
        return {}
    if filename not in span_indices:
        index = defaultdict(list)
        for node in stock_files[relative]["nodes"]:
            index[(node["start"], node["end"])].append(node)
        span_indices[filename] = index
    return span_indices[filename]
all_kinds = Counter()
boundaries_by_file = defaultdict(set)
kind_codes = {name: int(code) for code,names in stock["syntax_kinds"].items() for name in names}
for row in rows[1:]:
    for boundary in row.get('failed_boundaries', []):
        kind = boundary['kind'].removeprefix('Kind')
        assert kind in kind_codes, boundary
        boundaries_by_file[row['file']].add((boundary['start'], boundary['end'], kind_codes[kind]))
for row in rows[1:]:
    relative = str(Path(row['file']).relative_to(root))
    witness = stock_files[relative]
    source = (root / relative).read_bytes()
    assert witness['sha256'] == hashlib.sha256(source).hexdigest(), relative
    coverage = row['speculative_coverage']
    assert coverage['source_bytes'] == len(source), relative
    assert coverage['unvisited_nodes'] == 0, (relative, coverage)
    for finding in row['findings']:
        assert finding['root_kind'] == finding['reason']
        assert finding['depth'] >= 0
        all_kinds[finding['kind']] += 1
        site_where = finding.get('site_where', finding.get('boundary_where', ''))
        site_file = site_where.rsplit(':', 2)[0] if site_where else ''
        if site_file.startswith(str(root) + '/'):
            boundaries = boundaries_by_file[site_file]
            by_span = stock_spans(site_file)
            if finding.get('site_kind'):
                span = (finding.get('site_start', 0), finding['site_end'])
            else:
                span = (finding.get('start', 0), finding['end'])
            candidates = by_span.get(span, [])
            if finding.get('site_kind'):
                native_kind = finding['site_kind'].removeprefix('Kind')
                candidates = [node for node in candidates if native_kind == node['kind'] or
                              native_kind in stock.get('syntax_kinds', {}).get(str(node.get('kind_code', -1)), [])]
            depths = {sum((*span, kind) in boundaries for span,kind in zip(node["ancestors"],node["ancestor_kinds"],strict=True)) for node in candidates}
            if depths:
                assert finding['depth'] in depths, (relative, finding, depths)
                depth_matches += 1
            else:
                depth_unmatched += 1
        else:
            depth_without_source += 1
        if finding['kind'] not in counts or not finding['where'].startswith(str(root) + '/'):
            continue
        key = tuple(finding[k] for k in ('kind', 'where', 'reason', 'text')) + (finding.get('site_where', ''), finding.get('site_kind', ''), finding.get('site_start', 0), finding.get('site_end', 0))
        if key in seen:
            continue
        seen.add(key)
        depth = min(4, finding['depth'])
        counts[finding['kind']][depth] += 1
        reasons[(finding['kind'], finding['root_kind'])][depth] += 1
    files.append(dict(file=relative,sha256=witness['sha256'],**coverage))
missing = [{"file":str(p.relative_to(root)),"bytes":p.stat().st_size,"reason":"non-TypeScript input; no lowering sites"}
           for p in sorted(root.rglob('*')) if p.is_file() and str(p.relative_to(root)) not in stock_files]
total_bytes = sum(p.stat().st_size for p in root.rglob('*') if p.is_file())
examined_bytes = sum(f['source_bytes'] for f in files)
typescript_bytes = sum(p.stat().st_size for p in root.rglob('*') if p.is_file() and p.suffix in ('.ts','.a'))
ranking = sorted(reasons, key=lambda key: (-sum(reasons[key].values()),key))[:20]
result = dict(base='a5630a90d05abe85670ba1e00bff3643a2742e53',
              mapper_fix='69501280a81259fb512edbb8dd0e52c6eb0d88c8',
              upstream='050880ce59e30b356b686bd3144efe24f875ebc8',
              observation_kinds=dict(all_kinds),
              counts={kind:[counts[kind][d] for d in range(5)] for kind in counts},
              top20=[dict(kind=kind,root_kind=reason,counts=[reasons[(kind,reason)][d] for d in range(5)]) for kind,reason in ranking],
              source_bytes=total_bytes,examined_bytes=examined_bytes,
              examined_share=examined_bytes/total_bytes,
              typescript_source_bytes=typescript_bytes,typescript_examined_share=examined_bytes/typescript_bytes,
              coverage_definition='Every TypeScript AST child visited; value/statement/signature/binding positions attempted. This is speculative examination, not successful production lowering.',
              independent_depth_matches=depth_matches,independent_depth_unmatched=depth_unmatched,
              independent_depth_without_source=depth_without_source,
              files=files,unexamined=missing)
output.mkdir(parents=True,exist_ok=True)
(output/'RESULT.json').write_text(json.dumps(result,indent=2)+'\n')
text = '''Built a census-only speculative continuation overlay with typed placeholders and failure depth.
Base: `a5630a90`, merging `69501280` into `ed6e2975`; census changes stay inside its overlay.
Measured TypeScript 6.0.3 on the pinned adapted compiler corpus; tables below are observations.
Nested depth control, no-stubs baseline, depth mutant and output isolation checks are recorded in evidence.
Coverage means speculative AST examination; native semantics and successful compilation are outside this census.

'''
text += '| Depth | NotYet | Refused |\n|---|---:|---:|\n'
for d in range(5):
    text += f'| {"4+" if d==4 else d} | {counts["NotYet"][d]:,} | {counts["Refused"][d]:,} |\n'
text += f'| Total | {sum(counts["NotYet"].values()):,} | {sum(counts["Refused"].values()):,} |\n\n'
text += '| Kind | Root kind (exact reason) | 0 | 1 | 2 | 3 | 4+ | Total |\n|---|---|---:|---:|---:|---:|---:|---:|\n'
for kind,reason in ranking:
    values=[reasons[(kind,reason)][d] for d in range(5)]
    text += f'| {kind} | {html.escape(reason, quote=False).replace(chr(124), "&#124;").replace(chr(10), "<br>")} | '+ ' | '.join(f'{v:,}' for v in values+[sum(values)])+' |\n'
text += f'\nExamined **{examined_bytes:,} / {typescript_bytes:,} TypeScript source bytes ({100*examined_bytes/typescript_bytes:.6f}%)**, in {len(files)} files. '
text += f'Including non-TypeScript inputs, that is **{examined_bytes:,} / {total_bytes:,} bytes ({100*examined_bytes/total_bytes:.6f}%)**. '
text += 'Every AST child in every measured TypeScript file was visited, including checker-rejected bodies. '
text += 'Comments and whitespace are included in those parsed source bytes. Structural and type syntax is examined by traversal and the refusal scanner; it is not passed to a value lowerer.\n\n'
text += 'Unexamined non-TypeScript inputs:\n\n'
for item in missing:
    text += f'- `{item["file"]}`: {item["bytes"]:,} bytes; {item["reason"]}.\n'
text += f'\nStock TypeScript independently matched the ancestry depth of {depth_matches:,} findings at their actual source AST identities, including dependency attempts; {depth_unmatched:,} site spans had no exact stock AST match; {depth_without_source:,} findings had no compiler-source AST identity. '
text += 'The small control proves exact depth 1 and 2 independently of the corpus ranking.\n'
other = {kind:count for kind,count in all_kinds.items() if kind not in counts}
text += '\nOther raw attempt records, excluded from the NotYet/Refused site tables: ' + (json.dumps(other, sort_keys=True) if other else 'none') + '. See method.md for their interpretation.\n'
(output/'README.md').write_text(text)
print(json.dumps({k:v for k,v in result.items() if k not in ('files','top20')},indent=2))
