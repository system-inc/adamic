// Exact source-anchor mapping; never fall back to a nearest line.
const fs=require('node:fs'), path=require('node:path');
const ts=require(process.env.SCOUT_TYPESCRIPT);
const [historical,newTree,output]=process.argv.slice(2);
const sites=JSON.parse(fs.readFileSync(historical,'utf8'));
const cache=new Map();
function source(file){if(!cache.has(file)){const p=path.join(newTree,file), text=fs.readFileSync(p,'utf8'),sf=ts.createSourceFile(p,text,ts.ScriptTarget.Latest,true),fns=[];function visit(n){if(ts.isFunctionLike(n)&&n.body)fns.push(n);ts.forEachChild(n,visit);}visit(sf);cache.set(file,{text,sf,fns,lines:text.split(/\r?\n/)});}return cache.get(file);}
const out=sites.map((s,i)=>{const {sf,fns,lines}=source(s.file), matches=[];for(const fn of fns.filter(fn=>(fn.name?.getText(sf)||'<anonymous>')===s.context.name)){const lo=sf.getLineAndCharacterOfPosition(fn.getStart(sf)).line,hi=sf.getLineAndCharacterOfPosition(fn.end).line;for(let l=lo;l<=hi;l++){if(lines[l].trim()===s.source_line.trim()){const delta=(lines[l].length-lines[l].trimStart().length)-(s.source_line.length-s.source_line.trimStart().length);matches.push({line:l+1,column:s.column+delta,function_line:lo+1,source_line:lines[l]});}}}
 const r={id:'S'+String(i+1).padStart(2,'0'),historical_where:s.where,historical_units:s.units.map(u=>({where:u.where,name:u.name})),file:s.file,function:s.context.name,historical_source_line:s.source_line,mapping:'unmapped',matches};
 if(matches.length===1){Object.assign(r,matches[0],{mapping:'exact source line in uniquely matched enclosing function',where:path.join(newTree,s.file)+':'+matches[0].line+':'+matches[0].column});}return r;});
fs.writeFileSync(output,JSON.stringify(out,null,2)+'\n');console.log(JSON.stringify({sites:out.length,mapped:out.filter(s=>s.where).length,unmapped:out.filter(s=>!s.where).map(s=>s.id)}));
