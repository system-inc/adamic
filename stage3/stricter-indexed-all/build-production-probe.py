#!/usr/bin/env python3
"""Build an isolated probe; leave the production loader and checker inputs intact."""
import argparse
import csv
import hashlib
import json
import subprocess
from pathlib import Path

parser=argparse.ArgumentParser(description=__doc__)
parser.add_argument('--tree',type=Path,required=True)
parser.add_argument('--scratch',type=Path,required=True)
args=parser.parse_args()
root=Path(__file__).resolve().parent
repo=root.parents[1]
scratch=args.scratch.resolve();scratch.mkdir(parents=True,exist_ok=True)
hashes=json.loads((root/'evidence/ledger-source-hashes.json').read_text())
assert len(hashes)==79
mismatches=[name for name,expected in hashes.items() if hashlib.sha256((args.tree/name).read_bytes()).hexdigest()!=expected]
if mismatches: raise SystemExit('dated compiler source hashes differ: '+str(mismatches))
(scratch/'source-identity.json').write_text(json.dumps({'roots':79,'mismatches':mismatches},indent=2)+'\n')
rows=list(csv.DictReader((root/'evidence/ledger-rows.csv').open()))
excluded=[row for row in rows if row['option'] in ['exactOptionalPropertyTypes','useUnknownInCatchVariables']]
assert len(excluded)==72
manifest=[{'id':row['id'],'site':{'file':str(args.tree.resolve()/row['file']),'line':int(row['line']),'column':int(row['column']),'code':int(row['code'][2:]),'options':[row['option']]}} for row in excluded]
(scratch/'exclusions.json').write_text(json.dumps(manifest,indent=2)+'\n')
source=(repo/'internal/load/load.go').read_text()
original=source
needle='func load(paths []string, overlay map[string]string) (*Program, error) {'
assert source.count(needle)==1
source=source.replace(needle,needle+'\n\treturn loadIndexedLedgerProbe(paths, overlay, nil)\n}\n\nfunc loadIndexedLedgerProbe(paths []string, overlay map[string]string, exclusions []OptionSite) (*Program, error) {')
needle='\tif len(diagnostics) > 0 {\n\t\tsort.Strings(diagnostics)'
assert source.count(needle)==1
source=source.replace(needle,'\tif exclusions != nil {\n\t\tif !loaded.projectOptions { return nil, fmt.Errorf("indexed ledger probe requires project-owned .ts sources") }\n\t\tdiagnostics, err = filterIndexedLedgerDiagnostics(diagnostics, sites, exclusions)\n\t\tif err != nil { return nil, err }\n\t}\n'+needle)
source+='\n'+(root/'probe/loader.go.txt').read_text()
(scratch/'load.go').write_text(source)
# A generated main lives inside the module to permit internal package imports.
main=root/'probe/.generated';main.mkdir(parents=True,exist_ok=True)
(main/'main.go').write_text((root/'probe/main.go.txt').read_text())
(scratch/'overlay.json').write_text(json.dumps({'Replace':{str(repo/'internal/load/load.go'):str(scratch/'load.go')}}))
(scratch/'overlay-provenance.json').write_text(json.dumps({'production_loader_sha256':hashlib.sha256(original.encode()).hexdigest(),'overlay_loader_sha256':hashlib.sha256(source.encode()).hexdigest(),'exclusion_ids':[row['id'] for row in excluded],'production_loader_modified':False},indent=2)+'\n')
subprocess.run(['gofmt','-w',str(scratch/'load.go'),str(main/'main.go')],check=True)
subprocess.run(['go','build','-overlay',str(scratch/'overlay.json'),'-o',str(scratch/'production-probe'),str(main)],cwd=repo,check=True)
print(scratch/'production-probe')
