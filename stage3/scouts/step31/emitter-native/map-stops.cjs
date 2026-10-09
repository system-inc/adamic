const fs=require('node:fs'),path=require('node:path');
const ts=require(process.env.STEP31_TYPESCRIPT);
const [slice,walk,tree,destination]=process.argv.slice(2);
const manifest=JSON.parse(fs.readFileSync(path.join(slice,'slice.json'),'utf8'));
const rows=JSON.parse(fs.readFileSync(path.join(walk,'stops.json'),'utf8'));
const edits=[];
for(const row of rows){
 const relative=path.relative(path.join(walk,'slice'),row.file);
 const original=fs.readFileSync(path.join(slice,relative),'utf8');
 let current=original;
 const previous=edits.filter(e=>e.file===row.file);
 for(const edit of previous)current=current.slice(0,edit.start)+edit.replacement+current.slice(edit.end);
 const sf=ts.createSourceFile(relative,current,ts.ScriptTarget.Latest,true);
 let position=sf.getPositionOfLineAndCharacter(row.line-1,row.column-1);
 for(const edit of [...previous].reverse()){
  if(position>=edit.start+edit.replacement.length)position+=edit.end-edit.start-edit.replacement.length;
  else if(position>=edit.start)throw new Error('synthetic placeholder stop');
 }
 const declaration=manifest.declarations.find(d=>d.file===relative&&d.output_start<=position&&position<d.output_end);
 if(!declaration)throw new Error('not an audited source span');
 const upstream=fs.readFileSync(path.join(tree,relative),'utf8');
 const originalPosition=declaration.start+position-declaration.output_start;
 const upstreamAst=ts.createSourceFile(relative,upstream,ts.ScriptTarget.Latest,true);
 const coord=upstreamAst.getLineAndCharacterOfPosition(originalPosition);
 let nearest;
 function find(n){if(n.getStart(upstreamAst)<=originalPosition&&originalPosition<n.end&&ts.isFunctionLike(n)&&n.body)nearest=n;ts.forEachChild(n,find);}
 find(upstreamAst);
 row.original={file:relative,line:coord.line+1,column:coord.character+1,declaration:declaration.names,spanSha256:declaration.sha256,function:nearest?.name?.getText(upstreamAst)||'<anonymous>',lineText:upstream.split(/\r?\n/)[coord.line]};
 const stub=path.join(walk,`stub-${String(row.order).padStart(2,'0')}.stdout`);
 if(fs.existsSync(stub))edits.push(JSON.parse(fs.readFileSync(stub,'utf8')));
}
fs.writeFileSync(destination,JSON.stringify(rows,null,2)+'\n');
