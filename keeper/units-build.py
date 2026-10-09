import subprocess, re, json, collections, sys
repo = sys.argv[1]; times_path = sys.argv[2]; out = sys.argv[3]
rev = subprocess.check_output(['git','-C',repo,'rev-parse','origin/main'],text=True).strip()
# A testdata directory's _test.go files are side files copied into generated packages, never a package of their own.
files = [f for f in subprocess.check_output(['git','-C',repo,'ls-tree','-r','--name-only',rev],text=True).split() if f.endswith('_test.go') and 'testdata' not in f.split('/')]
latest = {}
for line in open(times_path):
    p = line.rstrip('\n').split('\t')
    if len(p) < 5 or p[0] == 'loom_seconds': continue
    try: latest[(p[3], p[4])] = float(p[0])
    except ValueError: pass
bypkg = collections.defaultdict(list)
for f in files:
    src = subprocess.check_output(['git','-C',repo,'show',f'{rev}:{f}'],text=True,errors='replace')
    tests = re.findall(r'^func (Test[A-Za-z0-9_]*)\(\w+ \*testing\.T\)', src, re.M)
    pkg = f.rsplit('/',1)[0]
    if tests: bypkg[pkg].append((f, tests))
def family(t):
    for pattern in (r'^(Test[A-Za-z0-9]+?)(_\d+|Union)$', r'^(Test[A-Za-z]+?)Unit\d+$', r'^(Test[A-Za-z0-9]+?)_.*_[0-9a-f]{12}$'):
        m = re.match(pattern, t)
        if m: return m.group(1)
    return t
def secs(pkg, tests):
    s = 0.0
    for t in tests: s += latest.get(('github.com/system-inc/adamic/'+pkg, t), 0.0)
    return round(s)
big = {'internal/oracle','internal/lower','internal/native'}
units = []
for pkg in sorted(bypkg):
    entries = bypkg[pkg]
    if pkg not in big:
        tests = [t for _,ts in entries for t in ts]
        units.append({'package':pkg,'files':[f for f,_ in entries],'tests':tests})
        continue
    # big: chunk files alphabetically by name stem, about 10 files each, never splitting a stem
    def stem(f):
        b = f.rsplit('/',1)[1][:-8]
        return re.split(r'_\d|_shard|_part', b)[0].split('_')[0]
    groups = collections.OrderedDict()
    for f, ts in sorted(entries):
        groups.setdefault(stem(f), []).append((f, ts))
    cur = []
    for g in groups.values():
        if cur and len(cur) + len(g) > 10:
            units.append({'package':pkg,'files':[f for f,_ in cur],'tests':[t for _,ts in cur for t in ts]}); cur = []
        cur += g
    if cur: units.append({'package':pkg,'files':[f for f,_ in cur],'tests':[t for _,ts in cur for t in ts]})
maxRows = 15
split = []
for u in units:
    perFile = [(f, [t for t in u['tests'] if t in set(ts)]) for f, ts in bypkg[u['package']] if f in u['files']]
    cur, curFamilies = [], set()
    for f, ts in perFile:
        fileFamilies = {family(t) for t in ts}
        if cur and len(curFamilies | fileFamilies) > maxRows:
            split.append({'package':u['package'],'files':[x for x,_ in cur],'tests':[t for _,y in cur for t in y]}); cur, curFamilies = [], set()
        cur.append((f, ts)); curFamilies |= fileFamilies
    if cur: split.append({'package':u['package'],'files':[x for x,_ in cur],'tests':[t for _,y in cur for t in y]})
units = split
for i,u in enumerate(units, 1):
    fam = collections.OrderedDict()
    for t in u['tests']: fam.setdefault(family(t), []).append(t)
    u['unit'] = f"u{i:03d}"
    u['rows'] = len(fam)
    u['test_functions'] = len(u['tests'])
    u['loom_seconds'] = secs(u['package'], u['tests'])
    u['slug'] = (u['package'].replace('/','-') + ('-' + u['files'][0].rsplit('/',1)[1][:-8] if sum(x['package']==u['package'] for x in units) > 1 else ''))[:60]
json.dump({'origin_main':rev,'units':units}, open(out,'w'), indent=1)
print(rev[:10], 'units', len(units), 'rows', sum(u['rows'] for u in units), 'test functions', sum(u['test_functions'] for u in units))
