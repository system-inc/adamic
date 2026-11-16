// Syntax/symbol inventory only. No reach inference.
const ts = require('../../api/node_modules/typescript');
const fs = require('fs'), path = require('path');
const root = process.argv[2];
function files(dir) { return fs.readdirSync(dir,{withFileTypes:true}).flatMap(e=>e.isDirectory()?files(path.join(dir,e.name)):e.name.endsWith('.ts')?[path.join(dir,e.name)]:[]); }
const roots=files(path.join(root,'src/compiler')).sort();
const program=ts.createProgram(roots,{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.NodeNext,moduleResolution:ts.ModuleResolutionKind.NodeNext,skipLibCheck:true});
const checker=program.getTypeChecker();
const unwrap=n=>{while(n && (ts.isParenthesizedExpression(n)||ts.isAsExpression(n)||ts.isNonNullExpression(n))) n=n.expression; return n;};
function saved(n,seen=new Set()) { n=unwrap(n); if(!n) return; if(ts.isPropertyAccessExpression(n)) return n; if(!ts.isIdentifier(n)) return; const s=checker.getSymbolAtLocation(n); if(!s||seen.has(s))return;seen.add(s); for(const d of s.declarations||[])if(ts.isVariableDeclaration(d)&&d.initializer){const r=saved(d.initializer,seen);if(r)return r;} }
const rows=[];
for(const file of roots){const sf=program.getSourceFile(file); function visit(n){ if(ts.isBinaryExpression(n)&&n.operatorToken.kind===ts.SyntaxKind.EqualsToken){const l=unwrap(n.left), r=saved(n.right), original=/^(?:\w+\.)?original[A-Z]\w*$/.test(unwrap(n.right).getText(sf));if(l&&ts.isPropertyAccessExpression(l)&&(r||original)){const s=checker.getSymbolAtLocation(l.name);const ds=s?.declarations||[];const declaredMethod=ds.some(d=>ts.isMethodSignature(d)&&(ts.isInterfaceDeclaration(d.parent)||ts.isTypeLiteralNode(d.parent))); const callable=checker.getNonNullableType(checker.getTypeAtLocation(l)).getCallSignatures().length>0; if(declaredMethod || (original && callable)) {const loc=n=>({file:path.relative(root,n.getSourceFile().fileName),line:n.getSourceFile().getLineAndCharacterOfPosition(n.getStart()).line+1,text:n.getSourceFile().text.split(/\r?\n/)[n.getSourceFile().getLineAndCharacterOfPosition(n.getStart()).line].trim()});rows.push({declaredMethod,site:loc(n),saved:r?loc(r):null,written:l.expression.getText(sf),read:r?r.expression.getText(r.getSourceFile()):null,member:l.name.text,savedMember:r?r.name.text:null});}}}ts.forEachChild(n,visit);}visit(sf);}
console.log(JSON.stringify({roots:roots.map(f=>path.relative(root,f)),rows},null,2));
