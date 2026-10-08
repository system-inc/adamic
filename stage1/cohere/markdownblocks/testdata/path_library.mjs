import fs from 'node:fs';
import vm from 'node:vm';
let bundle=fs.readFileSync(process.argv[2]+'/standalone.js','utf8');
const anchor='},tn=Mt;';
if(bundle.split(anchor).length!==2)throw new Error('pinned AstPath anchor');
bundle=bundle.replace(anchor,'},tn=Mt;globalThis.adamicAstPath=Mt;');
const context={module:{exports:{}},exports:{}};
vm.runInNewContext(bundle,context);
const Original=context.adamicAstPath;
const encode=text=>text.replace(/[\\\n\r\t]/g,c=>c==='\\'?'\\\\':c==='\n'?'\\n':c==='\r'?'\\r':'\\t');
let nodes=[],children=[],ids=new Map(),arrays=new Map();
const id=node=>ids.get(node)??-1;
const frame=value=>ids.has(value)?'N'+id(value):Array.isArray(value)?'A'+(arrays.get(value)??-1):typeof value==='number'?'I'+value:typeof value==='string'?'P'+value:'U';
const even=node=>ids.has(node)&&id(node)%2===0;
const some=(value,name,index)=>ids.has(value)&&(name===null||name===undefined||typeof name==='string')&&(index===null||index>=-1);
const never=(value,name,index)=>ids.has(value)&&typeof name==='string'&&index===-9;
const snapshot=path=>[path.stack.map(frame).join(','),frame(path.node),path.key??'',path.index??-1,frame(path.getName()),id(path.node),id(path.parent),id(path.grandparent),id(path.root),id(path.next),id(path.previous),path.isRoot,path.isInArray,path.isFirst,path.isLast,path.ancestors.filter(node=>ids.has(node)).map(id).join(','),[0,1,2,3,4].map(n=>id(path.getNode(n))).join(','),id(path.getParentNode(1)),id(path.findAncestor(even)),path.hasAncestor(even),path.match(),path.match(undefined),path.match(some,undefined),path.match(some,some),path.match(never)].join('|');
const identity=(path,index,array)=>(id(path.node)+1)*1000000+(index+1)*1000+arrays.get(array);
const observe=path=>{
 const rows=[snapshot(path),path.call(snapshot,'children'),path.call(snapshot,'children',-1),path.call(snapshot,'children',999999),path.call(snapshot,'absent')];
 for(let count=0;count<Math.min(3,path.ancestors.length);count++)rows.push(path.callParent(snapshot,count),''+path.callParent(p=>id(p.node),count),''+path.callParent(p=>p.isFirst,count));
 rows.push(''+path.call(p=>id(p.node),'children',0),snapshot(path),path.map(identity,'children').join(','));
 const visits=[];path.each((p,i,a)=>visits.push(identity(p,i,a)),'children');rows.push(visits.join(','),snapshot(path));return rows.join('\n');
};
if(process.argv[3]==='--key-gap'){
 const path=new Original({type:'root',children:[]});
 path.call(p=>process.stdout.write('key='+p.key+' type='+typeof p.key+'\n'),'absent','children',0);
 process.exit(0);
}
// Each context keeps its existing encoded row, without a corpus-sized V8 string.
let wrote=false;
for(const line of fs.readFileSync(process.argv[3],'utf8').split('\n')) {
 const fields=line.split('\t');
 if(fields[0]==='A'){nodes=[];children=[];ids=new Map();arrays=new Map()}
 else if(fields[0]==='N'){
  const node={type:fields[1]};ids.set(node,nodes.length);nodes.push(node);
  if(Number(fields[3])&2)node.children=[];
  children.push(fields[14]===''?[]:fields[14].split(',').map(Number));
 }else if(fields[0]==='F'){
  for(let i=0;i<nodes.length;i++)if(nodes[i].children){nodes[i].children=children[i].map(i=>nodes[i]);arrays.set(nodes[i].children,i)}
  const path=new Original(nodes[0]),rows=[];
  const walk=path=>{rows.push((path.node.children?.length??0)>0||path.isRoot?observe(path):snapshot(path));if(path.node.children)path.each(walk,'children')};walk(path);
  fs.writeSync(1,encode(rows.join('\n'))+'\n');
  wrote=true;
 }else if(line!=='')throw new Error('unknown path fixture');
}
if(!wrote)fs.writeSync(1,'\n');
