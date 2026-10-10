"""Independent piece provenance, coverage, typed-depth union and whole-file comparison.

Usage: check_piece_union.py ROOT PLAN RUN WHOLE_RECORD STOCK_JSON OUTPUT
Exit 0 means exact union parity; exit 2 means published semantic differences.
The dropped/duplicated/moved-piece mutants must fail the structural union audit.
"""
from collections import Counter, defaultdict
from copy import deepcopy
import hashlib
import json
from pathlib import Path
import sys

root, plan_path, run, baseline_path, stock_path, output = (Path(p).resolve() for p in sys.argv[1:7])
output.mkdir(parents=True, exist_ok=True)
plan = json.loads(plan_path.read_text())
plan_digest = hashlib.sha256(plan_path.read_bytes()).hexdigest()
identity = json.loads((run / 'INPUT.json').read_text())
assert identity['plan_sha256'] == plan_digest
for source in identity['sources']:
    assert hashlib.sha256((root / source['file']).read_bytes()).hexdigest() == source['sha256']
stock = json.loads(stock_path.read_text())
assert stock['typescript'] == plan['typescript'] == '6.0.3'
codes = {name:int(code) for code,names in stock['syntax_kinds'].items() for name in names}
indices = {}
stock_file = None
for file in stock['files']:
    assert hashlib.sha256((root / file['file']).read_bytes()).hexdigest() == file['sha256']
    index = defaultdict(list)
    for node in file['nodes']:
        index[(node['start'],node['end'],node['kind_code'])].append(node)
    indices[str(root / file['file'])] = index
    if str(root / file['file']) == plan['file']:
        stock_file = file
assert stock_file and stock_file['sha256'] == plan['sha256']
# Independently recover the indivisible statements from stock ancestor chains.
source_kind, function_kind, block_kind = codes['SourceFile'], codes['FunctionDeclaration'], codes['Block']
expected_atoms = set()
wrapper = None
for node in stock_file['nodes']:
    if node['kind_code'] == function_kind and len(node['ancestors']) == 1:
        # Planner records wrapper membership, checked separately against its source spelling.
        if any(a['parent'] == 'createTypeChecker' and node['start'] <= a['start'] < a['end'] <= node['end'] for a in plan['atoms']):
            wrapper = node
for node in stock_file['nodes']:
    if len(node['ancestors']) == 1 and node['ancestor_kinds'] == [source_kind] and node['kind_code'] != codes['EndOfFileToken']:
        if node is not wrapper:
            expected_atoms.add((node['start'],node['end'],node['kind_code']))
    if wrapper and len(node['ancestors']) == 3 and node['ancestor_kinds'] == [source_kind,function_kind,block_kind] and node['ancestors'][1] == [wrapper['start'],wrapper['end']]:
        expected_atoms.add((node['start'],node['end'],node['kind_code']))
if plan.get('version') == 2:
    # Recover each recursive cut from stock parent chains, not planner weights.
    nodes = stock_file['nodes']
    by_span_kind = {(n['start'],n['end'],n['kind_code']):n for n in nodes}
    split_functions = {(n['start'],n['end'],function_kind) for n in plan['split_functions']}
    split_containers = {(n['start'],n['end'],codes['ModuleDeclaration']) for n in plan['split_containers']}
    cuts = split_functions | split_containers
    assert cuts <= by_span_kind.keys(), 'invented recursive envelope'
    atom_spans = {(a['start'],a['end'],codes[a['kind']]):a for a in plan['atoms']}
    if plan.get('root_name'):
        roots = [min(plan['atoms'],key=lambda a:a['start'])]
        assert roots[0]['name'] == plan['root_name']
        roots = [(a['start'],a['end'],codes[a['kind']]) for a in roots]
    else:
        roots = list(expected_atoms)
    expected_atoms = set()
    used_cuts = set()
    def expand(key):
        assert key in by_span_kind, 'unknown recursive root'
        expected_atoms.add(key)
        if key not in cuts:
            assert atom_spans[key]['bytes'] == key[1]-key[0], 'incorrect leaf byte size'
            return
        used_cuts.add(key)
        parent = by_span_kind[key]
        depth = len(parent['ancestors'])
        children = [n for n in nodes if len(n['ancestors']) == depth + 2
                    and n['ancestors'][depth] == [parent['start'],parent['end']]
                    and n['ancestor_kinds'][depth] == parent['kind_code']
                    and n['ancestor_kinds'][depth+1] == (block_kind if key in split_functions else codes['ModuleBlock'])
                    and (key not in split_functions or n['kind_code'] == function_kind)]
        assert children, 'empty recursive cut'
        own = parent['end']-parent['start']-sum(n['end']-n['start'] for n in children)
        assert atom_spans[key]['bytes'] == own, 'incorrect recursive own-statement size'
        for child in children:
            expand((child['start'],child['end'],child['kind_code']))
    for key in roots:
        expand(key)
    assert used_cuts == cuts, 'unused recursive cut'
    assert all(p['bytes'] == sum(atom_spans[(a['start'],a['end'],codes[a['kind']])]['bytes'] for a in plan['atoms'] if a['id'] in p['atoms'])
               and p['bytes'] <= plan['threshold_bytes'] for p in plan['pieces']), 'piece exceeds threshold or size changed'
assert {(a['start'],a['end'],codes[a['kind']]) for a in plan['atoms']} == expected_atoms, 'planner omitted or invented source statements'
planned_atoms = {a['id']:a for a in plan['atoms']}
assigned = [atom for piece in plan['pieces'] for atom in piece['atoms']]
assert len(assigned) == len(set(assigned)) and set(assigned) == set(planned_atoms), 'plan is not an exact atom partition'
# Attribute each stock AST node to its nearest planned atom. Recursive roots
# overlap spatially, but the owned node partition must still count each once.
atom_keys = {(a['start'],a['end'],codes[a['kind']]):a['id'] for a in plan['atoms']}
owned_nodes = Counter()
for node in stock_file['nodes']:
    chain = [(node['start'],node['end'],node['kind_code'])]
    chain.extend((*span,kind) for span,kind in reversed(list(zip(node['ancestors'],node['ancestor_kinds'],strict=True))))
    owner = next((atom_keys[key] for key in chain if key in atom_keys),None)
    if owner is not None:
        owned_nodes[owner] += 1
assert set(owned_nodes) == set(planned_atoms), 'empty or missing owned AST atom'
rows = []
header = None
for piece in plan['pieces']:
    path = run / 'records' / f"piece-{piece['id']:03d}.jsonl"
    assert path.exists(), f"piece {piece['id']} did not complete"
    assert hashlib.sha256(path.read_bytes()).hexdigest() == path.with_suffix('.sha256').read_text().strip()
    raw = [json.loads(line) for line in path.read_text().splitlines()]
    assert len(raw) == 2
    if header is None:
        header = raw[0]
    assert header == raw[0], 'project checker header changed'
    rows.append(raw[1])
baseline_raw = [json.loads(line) for line in baseline_path.read_text().splitlines()]
assert len(baseline_raw) == 2 and baseline_raw[0] == header
assert hashlib.sha256(baseline_path.read_bytes()).hexdigest() == baseline_path.with_suffix('.sha256').read_text().strip()
baseline = baseline_raw[1]

def audit_union(records):
    ids = [r['piece']['id'] for r in records]
    assert len(ids) == len(set(ids)), 'duplicate piece in union'
    assert set(ids) == {p['id'] for p in plan['pieces']}, 'piece dropped from union'
    for record in records:
        piece = plan['pieces'][record['piece']['id']]
        assert record['file'] == plan['file']
        assert record['piece']['plan_sha256'] == plan_digest and record['piece']['atoms'] == piece['atoms']
        assert record['speculative_coverage']['unvisited_nodes'] == 0
        assert record['speculative_coverage']['total_nodes'] == sum(owned_nodes[a] for a in piece['atoms']), 'owned AST coverage double-counted or dropped nodes'
        assert record['speculative_coverage']['source_bytes'] == plan['source_bytes'], 'piece source byte identity changed'
        expected_units = {(planned_atoms[a]['start'], planned_atoms[a]['end'], codes[planned_atoms[a]['kind']]) for a in piece['atoms']}
        assert len(record['units']) == len(expected_units)
        allowed = {'prepass', *piece['atoms']}
        origins = {a:None for a in piece['atoms']}
        for atom, unit in zip(sorted(piece['atoms'],key=lambda a:planned_atoms[a]['start']),record['units'],strict=True):
            assert unit['kind'] == 'Kind' + planned_atoms[atom]['kind']
            origins[atom] = unit['where']
        for finding in record['findings']:
            atom = finding.get('piece_atom')
            assert atom in allowed, 'site moved between pieces: atom provenance violates partition'
            if atom != 'prepass':
                assert finding['unit'] == origins[atom], 'site moved between pieces: unit provenance changed'
            else:
                assert finding['phase'] == 'refusal_scan', 'body finding relabelled as prepass'
    return True

audit_union(rows)
mutants = {}
def caught(name, mutant, expected):
    try:
        audit_union(mutant)
    except AssertionError as error:
        assert expected in str(error), (name,error)
        mutants[name] = str(error)
    else:
        raise AssertionError(name + ' survived')
caught('piece dropped',deepcopy(rows[:-1]),'piece dropped')
caught('piece duplicated',deepcopy(rows)+[deepcopy(rows[0])],'duplicate piece')
mutant = deepcopy(rows)
source = next(r for r in mutant if any(f['piece_atom'] != 'prepass' for f in r['findings']))
target = next(r for r in mutant if r is not source)
finding = next(f for f in source['findings'] if f['piece_atom'] != 'prepass')
source['findings'].remove(finding)
target['findings'].append(finding)
caught('site moved between pieces',mutant,'site moved')
mutant = deepcopy(rows)
mutant[0]['speculative_coverage']['total_nodes'] += 1
caught('shifted owned AST count',mutant,'owned AST coverage')
(output / 'MUTANTS.json').write_text(json.dumps(mutants,indent=2)+'\n')

def finding_key(f):
    return tuple(f.get(k,0 if k in ('site_start','site_end') else '') for k in
                 ('kind','where','reason','text','site_where','site_kind','site_start','site_end'))

def combine(records):
    boundaries = defaultdict(set)
    for record in records:
        for filename, inventory in record.get('project_failed_boundaries',{}).items():
            for b in inventory:
                boundaries[filename].add((b['start'],b['end'],codes[b['kind'].removeprefix('Kind')]))
        for b in record.get('failed_boundaries',[]):
            boundaries[record['file']].add((b['start'],b['end'],codes[b['kind'].removeprefix('Kind')]))
    findings = {}
    changed = 0
    for record in records:
        for finding in record['findings']:
            f = deepcopy(finding)
            filename = f.get('site_where','').rsplit(':',2)[0]
            if filename in indices and f.get('site_kind'):
                candidates = indices[filename][(f.get('site_start',0),f['site_end'],codes[f['site_kind'].removeprefix('Kind')])]
                depths = {sum((*span,kind) in boundaries[filename] for span,kind in zip(n['ancestors'],n['ancestor_kinds'],strict=True)) for n in candidates}
                assert len(depths) == 1, ('stock AST depth ambiguity',f,depths)
                depth = depths.pop()
                changed += depth != f['depth']
                f['depth'] = depth
            key = finding_key(f)
            assert key not in findings or findings[key]['depth'] == f['depth']
            findings[key] = f
    return findings, boundaries, changed

whole, whole_boundaries, whole_changes = combine([baseline])
pieces, piece_boundaries, piece_changes = combine(rows)
missing = [whole[k] for k in sorted(whole.keys()-pieces.keys())]
added = [pieces[k] for k in sorted(pieces.keys()-whole.keys())]
depths = [dict(site=whole[k], piece_depth=pieces[k]['depth']) for k in sorted(whole.keys() & pieces.keys()) if whole[k]['depth'] != pieces[k]['depth']]
removed_boundaries, added_boundaries = [], []
for file in sorted(whole_boundaries.keys() | piece_boundaries.keys()):
    removed_boundaries.extend(dict(file=file,start=s,end=e,kind_code=k) for s,e,k in sorted(whole_boundaries[file]-piece_boundaries[file]))
    added_boundaries.extend(dict(file=file,start=s,end=e,kind_code=k) for s,e,k in sorted(piece_boundaries[file]-whole_boundaries[file]))
result = dict(file=plan['file'], baseline_sha256=hashlib.sha256(baseline_path.read_bytes()).hexdigest(),
              exact_match=not(missing or added or depths),
              owned_stock_nodes=sum(owned_nodes.values()), piece_walk_nodes=sum(r['speculative_coverage']['total_nodes'] for r in rows),
              whole_walk_nodes=baseline['speculative_coverage']['total_nodes'], whole_sites=len(whole), piece_sites=len(pieces),
              missing_sites=missing, added_sites=added, changed_depths=depths,
              removed_boundaries=removed_boundaries, added_boundaries=added_boundaries,
              raw_depths_recalibrated=dict(whole=whole_changes,pieces=piece_changes),
              mutants=mutants, runtimes=[dict(piece=p['id'],bytes=p['bytes'],**json.loads((run/f"piece-{p['id']:03d}.metrics.json").read_text())) for p in plan['pieces']])
(output / 'UNION.json').write_text(json.dumps(result,indent=2)+'\n')
print(json.dumps({k:v for k,v in result.items() if k not in ('missing_sites','added_sites','changed_depths','removed_boundaries','added_boundaries','runtimes')},indent=2))
print(json.dumps(dict(missing=len(missing),added=len(added),depth_changes=len(depths),removed_boundaries=len(removed_boundaries),added_boundaries=len(added_boundaries))))
sys.exit(0 if result['exact_match'] else 2)
