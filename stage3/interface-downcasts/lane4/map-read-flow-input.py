#!/usr/bin/env python3
"""Map stock UTF-16 cast spans to the shared adapter's UTF-8 AST positions."""
import json,pathlib,sys
root=pathlib.Path(sys.argv[1]);rows=json.loads(pathlib.Path(sys.argv[2]).read_text());out=[];sources={}
for row in rows:
 file=row['file']
 if file not in sources:sources[file]=(root/file).read_bytes().decode('utf-8').encode('utf-16-le')
 data=sources[file]
 assert data[2*row['start']:2*row['end']].decode('utf-16-le')==row['text']
 out.append({**row,'adapted':{'start':len(data[:2*row['start']].decode('utf-16-le').encode()),'end':len(data[:2*row['end']].decode('utf-16-le').encode()),'owner':'source-span mapping'}})
pathlib.Path(sys.argv[3]).write_text(json.dumps(out)+'\n')
print(f'PASS: {len(out)} exact cast spans mapped to UTF-8')
