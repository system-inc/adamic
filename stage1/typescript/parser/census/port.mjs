import {readFileSync} from 'node:fs';
import {registerHooks,stripTypeScriptTypes} from 'node:module';
registerHooks({resolve(s,c,next){if(s==='adamic')return {url:new URL('../../../../oracle/adamic.mjs',import.meta.url).href,shortCircuit:true};return next(s,c)},load(u,c,next){if(u.endsWith('.ts'))return {format:'module',source:stripTypeScriptTypes(readFileSync(new URL(u),'utf8'),{mode:'transform'}),shortCircuit:true};return next(u,c)}});
const {Parser}=await import('../parser.ts');
const {written}=await import('../nodes.ts');
const paths=readFileSync(process.argv[2],'utf8').trim().split('\n');
for(let i=0;i<paths.length;i++){
 const path=paths[i];console.log(`case ${i}`);
 try {
 const source=readFileSync(path,'utf8'), parser=new Parser(source,path), root=parser.file();
 const offsets=[0];let bytes=0;
 for(let j=0;j<source.length;j++){let cp=source.codePointAt(j);if(cp>65535){offsets.push(bytes);j++;}bytes+=cp<128?1:cp<2048?2:cp<65536?3:4;offsets.push(bytes);}
 for(const d of parser.diagnostics)console.log(`diagnostic ${d.code} ${offsets[d.start]} ${offsets[d.start+d.length]-offsets[d.start]} 1\t${written(d.message)}`);
 console.log('file');
 const lines=[];
 function walk(id,depth){const n=parser.nodes[id]; const flags=(n.optional?32:0)|(n.kind==='VariableDeclarationList'?Number(n.semantic):0);
 lines.push(`${depth} ${n.kind} ${offsets[n.pos]} ${offsets[n.end]} ${flags} ${n.literalFlags} ${n.list} ${+n.trailing} ${+n.multiLine}\t${n.operator}\t${written(n.text)}\t${written(n.raw)}\t${n.semantic}`);
 for(const child of n.children)walk(child,depth+1);}
 walk(root,0);process.stdout.write(lines.join('\n')+'\n');
 }catch(e){console.log('adapter-failure '+e.stack?.replaceAll('\n',' '));}
}
