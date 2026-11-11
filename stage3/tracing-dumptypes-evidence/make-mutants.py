from pathlib import Path
import shutil,json,hashlib
S=Path(__file__).resolve().parent;A=Path('/workspace/scratch/dp24r5m-after');rows=[]
for side,before,after in [('mutant-object','const objectFlags = getObjectFlags(type);','const objectFlags = (type as any).objectFlags;'),('mutant-intrinsic','intrinsicName: type.flags & TypeFlags.Intrinsic ? (type as IntrinsicType).intrinsicName : undefined,','intrinsicName: (type as any).intrinsicName,')]:
 T=Path('/workspace/scratch/dp24r5m-'+side);T.mkdir();shutil.copytree(A/'src',T/'src');f=T/'src/compiler/tracing.ts';text=f.read_bytes().decode();assert text.count(before)==1;f.write_bytes(text.replace(before,after).encode());files=sorted(p.relative_to(A/'src') for p in (A/'src').rglob('*') if p.is_file());changed=[str(p) for p in files if (A/'src'/p).read_bytes()!=(T/'src'/p).read_bytes()];assert changed==['compiler/tracing.ts'];rows.append(dict(side=side,sourceFiles=len(files),changed=changed,restored=after))
(S/'restoring-mutants-results.json').write_text(json.dumps(rows,indent=2)+'\n');print(json.dumps(rows))
