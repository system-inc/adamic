import sys
sys.dont_write_bytecode=True
from run import *
import shutil,hashlib
original=pathlib.Path('/workspace/scratch/sys-timer-property/D')
def ignore(directory,names):
 if pathlib.Path(directory)==original:return ['built','.git','node_modules']
 if pathlib.Path(directory)==original/'tests/baselines':return ['local']
 return []
shutil.copytree(original,S/'DP',ignore=ignore)
(S/'DP/node_modules').symlink_to(ROOT/'stage3/typescript/node_modules',target_is_directory=True)
assert (S/'DP/src/compiler/sys.ts').read_bytes()==(original/'src/compiler/sys.ts').read_bytes()
file='web-globals/timers.d.ts';published='    function clearTimeout(timeout: NodeJS.Timeout | string | number | undefined): void;';corrected='    function clearTimeout(timeout: unknown): void;'
packages=[ROOT/'internal/load/node_types/node_modules/@types/node',ROOT/'stage3/typescript/node_modules/@types/node']
for p in packages:
 text=(p/file).read_text();assert text.splitlines()[40]==published;assert text.count(published)==1;assert json.loads((p/'package.json').read_text())['version']=='25.3.3'
rows=json.loads((ROOT/'internal/load/host_axioms.json').read_text());assert not any(r['file']==file for r in rows)
row={'file':file,'line':41,'published':published,'corrected':corrected,'node':{'measurement_only':True,'stand_in_for':'Language host row; Runtime global timer implementation is outside this measurement'},'ruling':'Adamic, b5w3ycg, Nov 7 21:47Z, item 2: clearTimeout host input unknown; scratch D-prime stand-in only'}
rows.append(row);(S/'host_axioms.json').write_text(json.dumps(rows,indent=2)+'\n');(S/'timer-row.json').write_text(json.dumps(row,indent=2)+'\n')
overlay=json.loads(pathlib.Path('/workspace/scratch/sys-timer-stack/overlay/overlay.json').read_text());overlay['Replace'][str(ROOT/'internal/load/host_axioms.json')]=str(S/'host_axioms.json');(S/'overlay.json').write_text(json.dumps(overlay,indent=2)+'\n')
scope={}
for p in [ROOT/'internal/load/host_axioms.json',*[p/file for p in packages]]:scope[str(p)]=hashlib.sha256(p.read_bytes()).hexdigest()
(S/'original-host-inputs.json').write_text(json.dumps(scope,indent=2)+'\n')
text=pathlib.Path('/workspace/scratch/sys-timer-property/check.cjs').read_text()
view="""
const timerRow=require('./timer-row.json');const servedViews=[];
function timerHostView(file,text){if(text===undefined||!file.replaceAll('\\\\','/').endsWith('/node_modules/@types/node/'+timerRow.file))return text;const lines=text.split('\\n');const at=timerRow.line-1;assertTimer(lines[at]===timerRow.published,'published timer row drift '+file);assertTimer(text.split(timerRow.published).length===2,'timer declaration not unique '+file);lines[at]=timerRow.corrected;servedViews.push({file,line:timerRow.line,published:timerRow.published,corrected:timerRow.corrected});return lines.join('\\n');}
function assertTimer(ok,message){if(!ok)throw Error(message);}
"""
text=text.replace("function diag(d)",view+"\nfunction diag(d)")
text=text.replace("let text=read(f);if(!text)return text;","let text=timerHostView(f,read(f));if(!text)return text;")
text=text.replace("const builder=ts.createSolutionBuilder", "const solutionRead=solutionHost.readFile;solutionHost.readFile=f=>timerHostView(f,solutionRead(f));\nconst builder=ts.createSolutionBuilder")
text=text.replace("const adamic=ts.getPreEmitDiagnostics(p).map(diag);", "const adamic=ts.getPreEmitDiagnostics(p).map(diag);\nconst providerTypes=[];const cp=p.getTypeChecker();const sf=p.getSourceFile(root+'/src/compiler/sys.ts');function scanProvider(n){if(ts.isShorthandPropertyAssignment(n)&&n.name.text==='clearTimeout'){const symbol=cp.getShorthandAssignmentValueSymbol(n);const at=sf.getLineAndCharacterOfPosition(n.getStart(sf));providerTypes.push({line:at.line+1,column:at.character+1,type:cp.typeToString(cp.getTypeOfSymbolAtLocation(symbol,n))});}ts.forEachChild(n,scanProvider);}scanProvider(sf);")
text=text.replace('solutionExit,requireWitness}', 'solutionExit,requireWitness,providerTypes,servedViews}')
(S/'check.cjs').write_text(text)
assert run('census-build',['go','build','-buildvcs=false','-overlay='+str(S/'overlay.json'),'-o',S/'census','./stage3/census/latent/tool'])==0
print('prepared scratch host-axiom row and source-identical D-prime',flush=True)
