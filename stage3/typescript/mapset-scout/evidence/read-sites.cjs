// Read-only source extraction for this scout. Usage: node read-sites.cjs raw-entry.jsonl adapted-tree output.json
const fs = require('node:fs'), path = require('node:path');
const ts = require(process.env.SCOUT_TYPESCRIPT || '../../../api/node_modules/typescript');
const [raw, tree, output] = process.argv.slice(2);
const reason = "a Map or Set interface implemented by a value that isn't a Map or Set";
const records = fs.readFileSync(raw,'utf8').trim().split('\n').map(JSON.parse);
const events = records.flatMap(r => r.findings || []).filter(f => f.reason === reason);
const files = new Map(), declarations = new Map();
function walk(dir) { for (const e of fs.readdirSync(dir,{withFileTypes:true})) { const p=path.join(dir,e.name); if(e.isDirectory())walk(p); else if(p.endsWith('.ts')) { const s=ts.createSourceFile(p,fs.readFileSync(p,'utf8'),ts.ScriptTarget.Latest,true);files.set(p,s);function visit(n){if(ts.isFunctionDeclaration(n)&&n.name&&n.body){const loc=s.getLineAndCharacterOfPosition(n.getStart(s));const info={file:path.relative(tree,p),line:loc.line+1,name:n.name.text,text:n.getText(s)};const old=declarations.get(n.name.text);if(!old||n.modifiers?.some(m=>m.kind===ts.SyntaxKind.ExportKeyword))declarations.set(n.name.text,info);}ts.forEachChild(n,visit);}visit(s); } } }
walk(path.join(tree,'src/compiler'));
const sites=[...new Set(events.map(f=>f.where))].map(where=>{const es=events.filter(f=>f.where===where),m=/^(.*):(\d+):(\d+)$/.exec(where),s=files.get(m[1]),pos=s.getPositionOfLineAndCharacter(+m[2]-1,+m[3]-1);let fn,leaf;
function visit(n){if(n.getFullStart()>pos||n.end<=pos)return;leaf=n;if(ts.isFunctionLike(n)&&n.body)fn=n;ts.forEachChild(n,visit);}visit(s);
const line=s.text.split(/\r?\n/)[+m[2]-1];const identifier=/^[A-Za-z_$][\w$]*/.exec(line.slice(+m[3]-1))?.[0];const declaration=declarations.get(identifier);
const units=es.map(e=>{const r=records.find(r=>(r.units||[]).some(u=>u.where===e.unit));const u=r?.units.find(u=>u.where===e.unit);return {where:e.unit,name:u?.name||'',text:e.text};});
return {where,file:path.relative(tree,m[1]),line:+m[2],column:+m[3],events:es,units,source_line:line,identifier,declaration,context:fn?{name:fn.name?.getText(s)||'<anonymous>',line:s.getLineAndCharacterOfPosition(fn.getStart(s)).line+1,text:fn.getText(s)}:null};});
fs.writeFileSync(output,JSON.stringify(sites,null,2)+'\n');
