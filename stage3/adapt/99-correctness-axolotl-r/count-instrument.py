from pathlib import Path
import shutil
s=Path('/workspace/scratch/99r')
shutil.copytree(s/'before',s/'counted',symlinks=True)
p=s/'counted/src/compiler/checker.ts'
b=p.read_bytes()
changes={
 b'const typeParameters = concatenate(source.typeParameters, [source.thisType!]);':b'const typeParameters = concatenate((globalThis as any).__sortedCensus("14118", source.typeParameters), [source.thisType!]);',
 b'return type.resolvedTypeArguments;':b'return (globalThis as any).__sortedCensus("16980", type.resolvedTypeArguments);',
 b'context.inferredTypeParameters = concatenate(context.inferredTypeParameters, uniqueTypeParameters);':b'context.inferredTypeParameters = (globalThis as any).__sortedCensus("41574", concatenate(context.inferredTypeParameters, uniqueTypeParameters));',
 b'getUniqueTypeParameterName(concatenate(context.inferredTypeParameters, result), name)':b'getUniqueTypeParameterName((globalThis as any).__sortedCensus("41628", concatenate(context.inferredTypeParameters, result)), name)'
}
for old,new in changes.items():
 assert b.count(old)==1,(old,b.count(old))
 b=b.replace(old,new)
p.write_bytes(b)
