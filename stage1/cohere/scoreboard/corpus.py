#!/usr/bin/env python3
"""Inventory fetched source snapshots; no installs or repository scripts.
Default first-run scope: shortest nonempty .ts (excluding .d.ts) in each snapshot.
--all lists every tracked supported file; it never calls that the full quiet hundred.
"""
import argparse, hashlib, json, pathlib, subprocess
p=argparse.ArgumentParser();p.add_argument('fetch');p.add_argument('--all',action='store_true');p.add_argument('--out',required=True);a=p.parse_args()
rows=json.loads(pathlib.Path(a.fetch).read_text());files=[];inventory=[]
for row in rows:
    if row['status']!='fetched':raise SystemExit('unavailable snapshot: '+row['repo'])
    root=pathlib.Path(row['path'])
    actual=subprocess.check_output(['git','-C',str(root),'rev-parse','HEAD'],text=True).strip()
    if actual!=row['requested_sha']:raise SystemExit('wrong pin: '+row['repo'])
    tree=subprocess.check_output(['git','-C',str(root),'ls-tree','-r','-l','-z','HEAD']).decode().split('\0')
    blobs={}
    for entry in tree:
        if not entry:continue
        meta,path=entry.split('\t',1);parts=meta.split()
        if parts[1]=='blob':blobs[path]=int(parts[3])
    supported=[x for x in blobs if pathlib.Path(x).suffix in {'.ts','.tsx','.js','.jsx','.a','.json','.yaml','.yml','.gql','.graphql','.css','.scss','.md','.mdx'}]
    candidates=[x for x in supported if x.endswith('.ts') and not x.endswith('.d.ts') and blobs[x]>0]
    chosen=sorted(supported) if a.all else sorted(candidates,key=lambda x:(blobs[x],x))[:1]
    if not chosen:raise SystemExit('empty source snapshot: '+row['repo'])
    records=[]
    for path in chosen:
        data=(root/path).read_bytes();files.append(str(root/path));records.append(dict(path=path,bytes=len(data),sha256=hashlib.sha256(data).hexdigest()))
    inventory.append(dict(repo=row['repo'],sha=actual,tracked_supported_files=len(supported),selected_files=len(chosen),sources=records))
manifest=dict(snapshots=[dict(repository=row['repo'],root=row['path'],sha=row['sha']) for row in rows],name='23 public snapshots: '+('all tracked supported files' if a.all else 'shortest nonempty TS in each; first-run probes'),files=files)
out=pathlib.Path(a.out);out.write_text(json.dumps(manifest,indent=2)+'\n');out.with_suffix('.sources.json').write_text(json.dumps(inventory,indent=2)+'\n')
