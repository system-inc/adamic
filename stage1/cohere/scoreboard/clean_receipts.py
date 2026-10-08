#!/usr/bin/env python3
"""Write cleaned copies of host trace-corrupted receipts. Never replace originals.
Measured stdout, exit code, timing, source hashes and census remain byte-exact.
"""
import gzip, hashlib, json, pathlib, re, sys
root=pathlib.Path(sys.argv[1]);destination=pathlib.Path(sys.argv[2]);destination.mkdir(parents=True,exist_ok=True)
log=destination/'trace-cleanup.jsonl';header=re.compile(rb'(?<!\\)"stderr":');closing=re.compile(rb'(?<!\\)(?:\\\\)*"')
for path in sorted(root.glob('receipts-*.jsonl.gz')):
    if len(sys.argv)>3 and not path.name.startswith('receipts-'+sys.argv[3]+'-'):continue
    output=destination/path.name
    if output.exists():continue
    temporary=output.with_suffix(output.suffix+'.clean.tmp')
    digest=hashlib.sha256()
    with path.open('rb') as raw:
        for chunk in iter(lambda:raw.read(1024*1024),b''):digest.update(chunk)
    fields=0;inside=False;buffer=b'';okay=True
    try:
        with gzip.open(path,'rb') as source, gzip.open(temporary,'wb',compresslevel=6) as target:
            for chunk in iter(lambda:source.read(256*1024),b''):
                buffer+=chunk
                while buffer:
                    if inside:
                        match=closing.search(buffer)
                        if match:buffer=buffer[match.end():];inside=False;continue
                        end=len(buffer)
                        while end and buffer[end-1]==92:end-=1
                        buffer=buffer[end:];break
                    match=header.search(buffer)
                    if not match:
                        if len(buffer)>32:target.write(buffer[:-32]);buffer=buffer[-32:]
                        break
                    if len(buffer)<=match.end()+1:break
                    if buffer[match.end()]!=34:raise ValueError('stderr is not a JSON string')
                    target.write(buffer[:match.start()]);target.write(b'"stderr":""');fields+=1;buffer=buffer[match.end()+1:];inside=True
            if inside:raise ValueError('unterminated trace string')
            target.write(buffer)
    except (EOFError,OSError,ValueError) as error:
        temporary.unlink(missing_ok=True);okay=False;print('incomplete original retained',path.name,str(error),flush=True)
    if okay:
        temporary.rename(output)
        with log.open('a') as out:out.write(json.dumps(dict(file=path.name,original_sha256=digest.hexdigest(),original_bytes=path.stat().st_size,cleaned_bytes=output.stat().st_size,host_trace_fields_omitted=fields))+'\n')
        print(path.name,path.stat().st_size,output.stat().st_size,fields,flush=True)
