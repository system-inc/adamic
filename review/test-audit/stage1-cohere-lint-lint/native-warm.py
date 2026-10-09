from run import run
for name,pattern in [('TestNodeTableIsLinkOnlyFamily','^TestNodeTableIsLinkOnly_[0-9]{3}$'),('TestOwnedWitnesses','^TestOwnedWitnesses(_Setup|Union|_[0-9]{3})$')]:
 for i in range(1,4):
  r=run('warm-timing-'+name+'-'+str(i),pattern)
  if r['exit']!=0:raise SystemExit('native baseline not green')
