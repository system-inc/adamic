const fs=require('fs'),path=require('path'),ts=require('/workspace/adamic/stage3/api/node_modules/typescript');
const out=__dirname,base='/workspace/c3b22184-adapted/src/compiler';
const rows=[1,2,3,4].flatMap(i=>JSON.parse(fs.readFileSync(path.join(out,`group-${i}-rows.json`),'utf8')));if(rows.length!==79)throw Error('expected 79 roots');
const groups=new Map();for(const row of rows){const where=(row.where||'-').split('/src/compiler/').at(-1),kind=row.kind,reason=row.reason||'-',key=JSON.stringify([where,kind,reason]);if(!groups.has(key))groups.set(key,{where,kind,reason,roots:[]});groups.get(key).roots.push(row.roots[0].split('/src/compiler/')[1]);}
const previous=JSON.parse(fs.readFileSync(path.join(out,'previous-grouped-first-stops.json'),'utf8'));const old=new Map(previous.flatMap(g=>g.roots.map(r=>[r,[g.where,g.kind,g.reason]])));
const anchors=JSON.parse(fs.readFileSync(path.join(out,'forecast-anchors.json'),'utf8'));
const excerpts=[],constructs=[],comparison=[];for(const g of groups.values()){
 let forecast=null,construct='lowers whole',record=null;
 if(g.kind!=='accepted'){
  const m=g.where.match(/^(.*):(\d+):(\d+)$/);if(!m)throw Error(g.where);const file=path.join(base,m[1]),source=fs.readFileSync(file,'utf8'),lines=source.split('\n'),sf=ts.createSourceFile(file,source,ts.ScriptTarget.Latest,true);
  const pos=sf.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1),nodes=[],functions=[];function visit(n){if(n.getStart(sf)<=pos&&pos<n.end){nodes.push(n);if(ts.isFunctionLike(n))functions.push(n);ts.forEachChild(n,visit);}}visit(sf);
  const scope=functions.at(-1),names=functions.map(n=>n.name?.getText(sf)).filter(Boolean),owner=names.at(-1)||null;
  const cast=[...nodes].reverse().find(n=>ts.isAsExpression(n)||ts.isTypeAssertionExpression(n));
  const expression=cast||[...nodes].reverse().find(n=>ts.isAsExpression(n)||ts.isTypeAssertionExpression(n)||ts.isPropertyAccessExpression(n)||ts.isCallExpression(n)||ts.isBinaryExpression(n)||ts.isFunctionDeclaration(n));
  const site=expression?.getText(sf)||nodes.at(-1)?.getText(sf)||lines[+m[2]-1];construct=(owner?owner+': ':'')+site.replace(/\s+/g,' ').slice(0,450);
  const sourceLine=lines[+m[2]-1];
  if(m[1]==='debug.ts'&&sourceLine.includes('Object.create(Object.prototype) as FlowNode'))forecast='debug.ts:537 cast';
  else if(m[1]==='debug.ts'&&sourceLine.includes('Object.create(Array.prototype) as NodeArray<Node>'))forecast='debug.ts:577 cast';
  else if(m[1]==='semver.ts'&&sourceLine.trim()===anchors.find(a=>a.forecast==='semver.ts:287').source_line.trim())forecast='semver.ts:287';
  else if(m[1]==='performanceCore.ts'&&sourceLine.trim()===anchors.find(a=>a.forecast==='performanceCore.ts:29').source_line.trim())forecast='performanceCore.ts:29';
  else if(m[1]==='performance.ts'&&sourceLine.trim()===anchors.find(a=>a.forecast==='performance.ts:189').source_line.trim())forecast='performance.ts:189';
  else if(names.includes('tryGetPerformanceHooks')&&sourceLine.includes('typeof performance.now'))forecast="tryGetPerformanceHooks performance.now presence test (Lower a3)";
  else if(names.includes('findAncestor')&&/result .*cannot be served by implementation result/.test(g.reason))forecast='findAncestor overload-result wall';
  else if(names.includes('serializePropertySymbolsForClassOrInterface')&&/result .*cannot be served by implementation result/.test(g.reason))forecast='serializePropertySymbolsForClassOrInterface overload-result wall';
  const start=scope?sf.getLineAndCharacterOfPosition(scope.getStart(sf)).line+1:+m[2],end=scope?sf.getLineAndCharacterOfPosition(scope.end).line+1:+m[2],whole=!!scope&&end-start+1<30;
  const text=whole?source.slice(sf.getPositionOfLineAndCharacter(start-1,0),scope.end):lines.slice(Math.max(0,+m[2]-4),+m[2]+3).join('\n');
  record={where:g.where,kind:g.kind,reason:g.reason,file,name:m[1].replaceAll('/','_')+'-'+m[2]+'-'+m[3]+'.txt',scope_name:owner,scope_names:names,scope_start:start,scope_end:end,whole_function:whole,text,site_text:site,source_line:sourceLine,construct};
  constructs.push(record);
 }
 g.forecast=forecast||(g.kind==='accepted'?'not applicable, lowers whole':'none, new');g.new_group=!forecast&&g.kind!=='accepted';g.construct=construct;g.count=g.roots.length;
 if(g.new_group){excerpts.push(record);fs.writeFileSync(path.join(out,record.name),record.text+'\n');}
 g.unchanged=0;g.moved=0;
 for(const root of g.roots){if(!old.has(root))throw Error('root absent from baseline '+root);const current=[g.where,g.kind,g.reason],unchanged=JSON.stringify(old.get(root))===JSON.stringify(current);g[unchanged?'unchanged':'moved']++;comparison.push({root,previous:old.get(root),current,change:unchanged?'unchanged':'moved',forecast:g.forecast,construct});}
 g.new_tuple=!previous.some(p=>p.where===g.where&&p.kind===g.kind&&p.reason===g.reason);g.comparison=g.unchanged===g.count?'unchanged':g.new_tuple&&g.new_group?'new':'moved';
}
const result=[...groups.values()].sort((a,b)=>b.count-a.count||a.where.localeCompare(b.where));for(const [name,value] of [['grouped-first-stops.json',result],['comparison.json',comparison],['constructs.json',constructs],['source-excerpts.json',excerpts]])fs.writeFileSync(path.join(out,name),JSON.stringify(value,null,2)+'\n');
console.log('roots='+rows.length+' groups='+result.length+' new_forecast_groups='+excerpts.length);
