from pathlib import Path
import json,shutil,difflib,hashlib
out=Path(__file__).resolve().parent;base=Path('/workspace/orm-adapted')
manifest=json.loads((out.parent/'overload-results-main/adapted-source-hashes.json').read_text())
assert all(hashlib.sha256((base/p).read_bytes()).hexdigest()==s for p,s in manifest.items())
# Copy only source/config inputs, never the checkout's large tests or installed dependencies.
for label in ['checked','typed-old']:
 target=Path('/workspace/cbp-'+label);target.mkdir(exist_ok=True)
 shutil.copytree(base/'src',target/'src',dirs_exist_ok=True)
 for name in ['package.json']:
  if (base/name).exists():shutil.copyfile(base/name,target/name)
 for name in ['builder.ts','builderPublic.ts']:
  p=target/'src/compiler'/name;s=p.read_bytes().decode();nl='\r\n' if '\r\n' in s else '\n'
  if label=='checked':
   if name=='builder.ts':
    # Both public-to-the-checkout overloads promise exactly the implementation's honest result.
    a='): SemanticDiagnosticsBuilderProgram;';b='): EmitAndSemanticDiagnosticsBuilderProgram;'
    assert s.count(a)==1 and s.count(b)==1
    s=s.replace(a,'): BuilderProgram | undefined;').replace(b,'): BuilderProgram | undefined;')
   else:
    for kind,typ,required in [('SemanticDiagnosticsBuilderProgram','SemanticDiagnosticsBuilderProgram',['getSemanticDiagnosticsOfNextAffectedFile']),('EmitAndSemanticDiagnosticsBuilderProgram','EmitAndSemanticDiagnosticsBuilderProgram',['getSemanticDiagnosticsOfNextAffectedFile','emitNextAffectedFile'])]:
     a='    return createBuilderProgram('+nl+'        BuilderProgramKind.'+kind+','
     assert a in s;s=s.replace(a,'    const builder = createBuilderProgram('+nl+'        BuilderProgramKind.'+kind+',',1)
     start=s.index('    const builder = createBuilderProgram('+nl+'        BuilderProgramKind.'+kind+',')
     end=s.index(nl+'    );',start)+len(nl+'    );')
     check='builder === undefined'+''.join(' || !("'+prop+'" in builder) || typeof builder.'+prop+' !== "function"' for prop in required)
     tail=nl+'    if ('+check+') {'+nl+'        throw new Error("Builder program does not have the requested kind");'+nl+'    }'+nl+'    return builder as '+typ+';'
     s=s[:end]+tail+s[end:]
  else:
   if name=='builder.ts':
    s=s.replace('export interface BuilderCreationParameters {','export interface BuilderCreationParameters<T extends BuilderProgram = BuilderProgram> {',1)
    a='    oldProgram: BuilderProgram | undefined;';assert s.count(a)==1;s=s.replace(a,'    oldProgram: T | undefined;',1)
    start=s.index('export function getBuilderCreationParameters(');end=s.index(nl+'function getTextHandlingSourceMapForSignature',start)
    block=s[start:end].replace('getBuilderCreationParameters(','getBuilderCreationParameters<T extends BuilderProgram = BuilderProgram>(').replace('oldProgramOrHost?: BuilderProgram | CompilerHost','oldProgramOrHost?: T | CompilerHost').replace('configFileParsingDiagnosticsOrOldProgram?: readonly Diagnostic[] | BuilderProgram','configFileParsingDiagnosticsOrOldProgram?: readonly Diagnostic[] | T').replace('): BuilderCreationParameters {','): BuilderCreationParameters<T> {').replace('let oldProgram: BuilderProgram;','let oldProgram: T;').replace(' as BuilderProgram;',' as T;')
    s=s[:start]+block+s[end:]
    for typ in ['SemanticDiagnosticsBuilderProgram','EmitAndSemanticDiagnosticsBuilderProgram']:
     a='kind: BuilderProgramKind.'+typ+','+nl+'    builderCreationParameters: BuilderCreationParameters,'
     assert a in s;s=s.replace(a,'kind: BuilderProgramKind.'+typ+','+nl+'    builderCreationParameters: BuilderCreationParameters<'+typ+'>,',1)
  p.write_bytes(s.encode())
  before=(base/'src/compiler'/name).read_bytes().decode()
  diff=''.join(difflib.unified_diff(before.splitlines(True),s.splitlines(True),fromfile='a/src/compiler/'+name,tofile='b/src/compiler/'+name))
  (out/(label+'-'+name+'.patch')).write_bytes(diff.encode())
 changes={}
 for p in sorted((target/'src').rglob('*.ts')):
  rel=str(p.relative_to(target));sha=hashlib.sha256(p.read_bytes()).hexdigest()
  if sha!=manifest.get(rel):changes[rel]=sha
 (out/(label+'-source-changes.json')).write_text(json.dumps(changes,indent=2)+'\n')
 print(label,'changed files:',list(changes))
(out/'adapted-source-hashes.json').write_text(json.dumps(manifest,indent=2)+'\n')
