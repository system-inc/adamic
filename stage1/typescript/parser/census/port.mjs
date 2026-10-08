import {readFileSync} from 'node:fs';
import {registerHooks,stripTypeScriptTypes} from 'node:module';
registerHooks({resolve(s,c,next){if(s==='adamic')return {url:new URL('../../../../oracle/adamic.mjs',import.meta.url).href,shortCircuit:true};return next(s,c)},load(u,c,next){if(u.endsWith('.ts'))return {format:'module',source:stripTypeScriptTypes(readFileSync(new URL(u),'utf8'),{mode:'transform'}),shortCircuit:true};return next(u,c)}});
const {Parser}=await import('../parser.ts');
const {written}=await import('../nodes.ts');
import {Worker,isMainThread,parentPort} from 'node:worker_threads';
function parse(path){const lines=[];let stage='parse';try{
 const source=readFileSync(path,'utf8'), parser=new Parser(source,path), root=parser.file();
 stage="adapter";const offsets=[0];let bytes=0;
 for(let j=0;j<source.length;j++){let cp=source.codePointAt(j);if(cp>65535){offsets.push(bytes);j++;}bytes+=cp<128?1:cp<2048?2:cp<65536?3:4;offsets.push(bytes);}
 for(const d of parser.diagnostics)lines.push(`diagnostic ${d.code} ${offsets[d.start]} ${offsets[d.start+d.length]-offsets[d.start]} 1\t${written(d.message)}`);
 lines.push('file');

 function walk(id,depth){const n=parser.nodes[id]; const flags=(n.optional?32:0)|(n.kind==='VariableDeclarationList'?Number(n.semantic):0);
 lines.push(`${depth} ${n.kind} ${offsets[n.pos]} ${offsets[n.end]} ${flags} ${n.literalFlags} ${n.list} ${+n.trailing} ${+n.multiLine}\t${n.operator}\t${written(n.text)}\t${written(n.raw)}\t${n.semantic}`);
 for(const child of n.children)walk(child,depth+1);}
 walk(root,0);return lines;

}catch(e){return [stage+'-failure '+e.stack?.replaceAll('\n',' ')];}}
if(!isMainThread){parentPort.on('message',path=>parentPort.postMessage(parse(path)));}
else if(process.argv[2]){
 const paths=readFileSync(process.argv[2],'utf8').trim().split('\n');let worker;
 function fresh(){return new Worker(new URL(import.meta.url),{stderr:true,stdout:true});}
 worker=fresh();
 for(let i=0;i<paths.length;i++){
  console.log(`case ${i}`);
  const rows=await new Promise(resolve=>{
   const active=worker;let stderr='';
   const onStderr=data=>{stderr=(stderr+data.toString()).slice(-1200)};
   active.stderr.on('data',onStderr);
   function finish(rows,restart=false){clearTimeout(timer);active.removeListener('message',onMessage);active.removeListener('exit',onExit);active.removeListener('error',onError);active.stderr.removeListener('data',onStderr);if(restart)worker=fresh();resolve(rows)}
   const onMessage=rows=>finish(rows);
   const onExit=code=>finish(['parse-failure worker exit '+code+' '+stderr.replaceAll('\n',' ')],true);
   const onError=error=>finish(['adapter-failure worker '+error.message],true);
   const timer=setTimeout(()=>{finish(['parse-failure timeout (5s)'],true);active.terminate()},5000);
   active.once('message',onMessage);active.once('exit',onExit);active.once('error',onError);
   active.postMessage(paths[i]);
  });
  process.stdout.write(rows.join('\n')+'\n');
 }
 await worker.terminate();
}
