import json,pathlib,re,subprocess
root=pathlib.Path.cwd();e=root/'review/compiler/bridge-products'
rows=json.loads((e/'results.json').read_text());by={r['label']:r for r in rows}
pairs=re.findall(r'func TestProduct_(\w+)\(t \*testing.T\) \{\s*t.Parallel\(\)\s*bridgeProduct\(t, bridgeRepository\(t\), "([^"]+)"\)',(root/'bridge/tsgo/product_units_test.go').read_text())
old=subprocess.check_output(['git','show','c9df51ed:bridge/tsgo/products_test.go'],text=True)
old_names=json.loads('['+re.search(r'var bridgeProductNames = \[\]string\{([^}]+)\}',old).group(1)+']')
assert set(old_names)=={p for _,p in pairs} and len(pairs)==19
manifest=[]
for path in pathlib.Path('/workspace/bridge-products-measured-cache').glob('*.inputs'):
 text=path.read_text();name=text.splitlines()[0].removeprefix('name bridge-test-')
 if name in old_names:
  manifest.append(dict(product=name,key=path.stem,inputs=text.splitlines()))
assert len(manifest)==19
(e/'product-inputs.json').write_text(json.dumps(sorted(manifest,key=lambda r:r['product']),indent=2)+'\n')
lines=(e/'merged-checks.builds.txt').read_text().splitlines()
products=[line for line in lines if line.split()[1].startswith('bridge-test-') and line.split()[1]!='bridge-test-bundle']
assert products and all(line.split()[3]=='hit' for line in products)
coverage=dict(old_products=old_names,new_products={product:'TestProduct_'+suffix for suffix,product in pairs},old_bridge_checks=36,new_bridge_checks=36,registered_bridge_checks_unchanged=subprocess.run(['git','diff','--quiet','c9df51ed','HEAD','--','bridge/tsgo/registered_units_test.go']).returncode==0,merged_product_fetches=len(products),merged_product_misses=0)
(e/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
s='''Nineteen bridge artifacts now have TestProduct units and a shared buildcache.Product recipe; the convention is documented.
The doc commit is a44c7eae, the test implementation is 73f480a9, and the main merges are 20f1ed30 and 140ebf83.
All nineteen products, nineteen input guards, sample bridge checks and measured compiler-corpus checks pass within 60 seconds.
Nineteen omitted-input mutants serve stale products and fail; removing one registered product fails the count assertion.
No production code, oracle fixtures, gate implementation or shared-store implementation changed; other compiler-corpus leaves were not rerun.

This serves step 79 and task #sykbpks: builds have their own units instead of making a bridge observation carry cold setup.
The first commit contains only the requested docs/test-products.md convention, preserving its words and restoring markdown.
All implementation changes are in _test.go. Every branch commit is the allowed doc, tests, review evidence or a merge of main.
Main b5245943 and cd0ecb04 were merged without conflicts; they changed unrelated estree, printer, regexp, lint and YAML tests.

The same bridgeProduct recipe serves all nineteen product tests and every bridge consumer. There are no sync.Once builders.
Every product test calls t.Parallel first, has no subtests and only fetches its product. Product keys contain sorted source,
C and assembly files, embedded files, module metadata and local module sums from the Go dependency graph. Discovery uses
Go list without -export, so a hit does not compile dependencies. External module sources are fingerprinted as named flags,
since buildcache.Files accepts repository-relative paths. No dependency source is copied.
Build flags include effective Go configuration, the archive and native modes, sanitizer and counting flags, compiler options,
bridge source recipes and build-related environment values. Toolchain names Go, clang and the effective C tools.
The shared full dependency graph is conservative: changing an unrelated bridge artifact can invalidate extra products.
The complete manifests and keys are in product-inputs.json. Go action caches remain keyed and validated by Go.

runBridgeCase fetches every required product before starting its existing budget clock. ADAMIC_UNIT_BUDGET=1 enables
that clock's budget on the reference instance; there is no unconditional deadline added. Existing subprocess hang guards remain.
The old nineteen artifacts and thirty-six bridge observations remain covered; coverage.json lists old products to new tests.
TestBridgeProductUnitsCoverEveryProduct counts nineteen product bindings and checks every case's declared products.
Its planted failure removes the native binding and fails with "19 products, 18 units, want 19".

For each product, its input guard uses that product's exact Inputs against a small test-owned source tree. It changes the
recipe content and fetches a marker product through the resulting key and buildcache.Get. Each overlay drops only
bridge/tsgo/products_test.go from one product's inputs. The second fetch serves "version one" instead of "version two"
and the corresponding guard fails with "served stale product". These are cache behavior failures, not compilation failures.
All nineteen unmutated guards pass separately on the final source; their seconds and every mutant result are below.

Cold measurements run each TestProduct alone in a fresh go test process, -count=1, -json, -timeout 90s, GOMAXPROCS=4,
ADAMIC_GATE_UNCACHED=1 and a fresh runtime cache (XDG_CACHE_HOME). Each product's own entry is absent: its log records
one miss for that product. Prerequisite products share the cache and are built in the listed order. The Go action cache
prepared by setup and earlier builds is retained; these are cold product and runtime caches, not an empty Go toolchain cache.
The initial archive run also built cold Go archive actions and took 56.510 seconds including process setup; its logs remain
under initial-. After the external-module key audit, every mutant and measurement was rerun. The final-source table follows.

| Product unit | Artifact | Cold seconds including setup | Input guard seconds | Omitted-input mutant |
| --- | --- | ---: | ---: | --- |
'''
for suffix,product in pairs:
 p=by['product-'+suffix];g=by['final-input-'+suffix];m=by['mutant-'+suffix]
 assert p['exit']==0 and g['exit']==0 and m['exit']==1 and p['wall_seconds']<60
 s+=f'| TestProduct_{suffix} | {product} | {p["wall_seconds"]:.3f} | {g["wall_seconds"]:.3f} | caught |\n'
s+='''
All sixteen active sample bridge leaves pass, including ABI, sanitizer, ownership, refusal and wrong-position checks.
On TypeScript 6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8, the oracle and first timing checker leaves pass.
Each consumer starts in a fresh Go test process. Every product census line in these consumer logs is a hit.
The two-leaf invocation uses the same Go test selection as a pooled unit; it does not rebuild a product.

| Consumer invocation | Seconds including process setup | Builds |
| --- | ---: | --- |
'''
for label in ['TestBridgeABI','TestBridgeOracleSample','two-fresh-leaves','corpus-TestBridgeOracleChecker','corpus-TestBridgeTimingRound1Checker','two-fresh-corpus-leaves']:
 r=by[label];s+=f'| {r["test"]} | {r["wall_seconds"]:.3f} | cache hits only |\n'
s+='\nSample leaf measurements:\n\n| Leaf | Seconds |\n| --- | ---: |\n'
for r in rows:
 if r['label'].startswith('consumer-'):s+=f'| {r["test"]} | {r["wall_seconds"]:.3f} |\n'
s+='''
Commands and exact output are in results.json, individual JSON logs, run.py and consumers.py.
The merged-tip check runs only TestProduct_, TestBridgeProductInput_, and the three coverage/cache tests;
all pass and all real product fetches are hits. No whole package tests or full gate ran. No fixture was added,
so internal/oracle/counts.md is unchanged. The other eighteen compiler-corpus leaf selections, other platforms,
a completely empty Go action cache and shared-store transport are not covered by this unit.

Setup: Go ready 0.025 s, Node ready 0.027 s, submodules ready 0.072 s, markdown dependencies 0.076 s,
clang ready 0.187 s, Go build ready 38.870 s, setup complete 39.166 s. setup.txt holds its exact lines.
nproc is 5; cpu.max is 400000/100000, a four-CPU quota. Go 1.27.1, clang 20.1.8, Node 24.19.0.
Old regenerable Go cache files and this unit's superseded products were cleared to provide disk space.

Mandated integration lane checks on the committed merged tip:

```text
'''+(e/'lane-checks.txt').read_text()+'''```
'''
(e/'report.md').write_text(s)
