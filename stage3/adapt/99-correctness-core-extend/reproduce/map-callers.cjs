const fs=require('fs'),path=require('path'),{TraceMap,originalPositionFor}=require('/workspace/adamic/stage3/typescript/node_modules/@jridgewell/trace-mapping');
const root='/workspace/scratch/core-extend',cache=new Map(),counts={},unmapped=[];
for(const folder of ['full-count'])for(const name of fs.readdirSync(root+'/'+folder+'/calls').filter(n=>n.endsWith('.jsonl'))){
for(const line of fs.readFileSync(root+'/'+folder+'/calls/'+name,'utf8').trim().split('\n')){
 if(!line)continue;const row=JSON.parse(line);let found=null;
 for(const frame of row.stack){
  const match=frame.match(/(?:\(|at )((?:file:\/\/)?[^()]+):(\d+):(\d+)\)?$/);if(!match)continue;
  const file=match[1].replace(/^file:\/\//,''),mapFile=root+'/counted/built/local/'+path.basename(file)+'.map';
  if(file.includes('/src/')&&file.endsWith('.ts')){const source=file.replace(/^.*\/src\//,'src/');const line=Number(match[2]);if(source==='src/compiler/core.ts'&&line>=1505&&line<=1530)continue;found=source+':'+line;break;}
  if(!fs.existsSync(mapFile))continue;
  let map=cache.get(mapFile);if(!map){map=new TraceMap(JSON.parse(fs.readFileSync(mapFile,'utf8')));cache.set(mapFile,map);}
  const pos=originalPositionFor(map,{line:Number(match[2]),column:Number(match[3])-1});if(!pos.source)continue;
  const source=pos.source.replace(/^.*\/src\//,'src/');
  if(source==='src/compiler/core.ts'&&pos.line>=1505&&pos.line<=1530)continue;
  found=source+':'+pos.line;break;
 }
 const key=row.owner+':'+(found||'unmapped');counts[key]=(counts[key]||0)+1;if(!found&&unmapped.length<10)unmapped.push(row);
}}
console.log(JSON.stringify({counts,unmappedSamples:unmapped,mapFiles:[...cache.keys()]},null,2));
