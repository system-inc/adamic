from audit_run import *
plan=json.load(open(P/'plan.json')); native=[r for r in rows if 'Oracle' in r or 'Timing' in r]+['TestBridgeRegion']; regular=[r for r in rows if r not in native and r not in ['TestBridgeABI'] and 'WrongPosition' not in r and not any(k in r for k in ['InputLength','OutputLength','StaleHandle','Linkage','OutputFree','RegionOwnership'])]
# The large-package rule permits a matrix over statically reached rows.
# Include all regular consumers of each adapter; witnesses receive their own weakened-check runs.
def evaluate(m):
 if m['id'] in ['P_ABI','P_CREATE','P_ABI_RELEASE','P_FREE','P_RESULT_FREE']:selected=['TestBridgeABI']
 elif m['id'] in ['M05','M06']:selected=['TestBridgeABI']+native
 elif m['id'] in ['M08','P_LOWER']:selected=regular+['TestTSGoRequiresLink']
 else:selected=native
 # Runtime selectors do not change compiled product bytes. Compiler selectors require separate product caches.
 cache=str(S/'cache'/m['id']) if m['id'] in ['M08','P_LOWER'] else None
 for corpus in [False,True]:
  chosen=[r for r in selected if (any(r.endswith(x) for x in ['Checker','Parser','Types','Utilities']))==corpus]
  if not chosen:continue
  # Split by file to bound metadata hashing and corpus subprocesses well below ninety seconds.
  for at in range(0,len(chosen),6):
   part=chosen[at:at+6];label=m['id']+'-'+str(corpus)+'-'+str(at)
   if (P/(label+'.json')).exists():
    old=json.load(open(P/(label+'.json')))
    if not old['cooked'] and not any('clang with checker archive failed' in e.get('Output','') for e in old['events']):continue
   d=run(label,'^('+'|'.join(part)+')$',corpus,m['id'],cache=cache)
   if d['cooked']:raise SystemExit('COOKED '+m['id'])

# Baseline measurements were serial. Matrix runs check correctness only and use independent process environments.
from concurrent.futures import ThreadPoolExecutor
with ThreadPoolExecutor(max_workers=3) as pool:
 list(pool.map(evaluate,plan))
