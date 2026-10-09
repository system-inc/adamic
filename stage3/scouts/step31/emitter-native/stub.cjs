// Discovery only. Preserve an explicit or checker-inferred signature, then throw.
const fs = require('node:fs');
const ts = require(process.env.STEP31_TYPESCRIPT);
const [file,line,column] = process.argv.slice(2);
const text=fs.readFileSync(file,'utf8');
const source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const position=source.getPositionOfLineAndCharacter(+line-1,+column-1);
const candidates=[];
function walk(node) {
    if(node.getStart(source)<=position && position<node.end && ts.isFunctionLike(node)&&node.body) candidates.push(node);
    ts.forEachChild(node,walk);
}
walk(source);
candidates.sort((a,b)=>a.end-a.pos-(b.end-b.pos));
const node=candidates[0];
if(!node) throw new Error('No enclosing function; stop the walk rather than invent an edit');
let annotation='';
if(!node.type && (ts.isFunctionDeclaration(node)||ts.isMethodDeclaration(node))) {
    const program=ts.createProgram([file],{strict:true,target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext});
    const checker=program.getTypeChecker();
    let same;
    function find(n){if(n.pos===node.pos&&n.kind===node.kind)same=n;ts.forEachChild(n,find);}
    find(program.getSourceFile(file));
    annotation=': '+checker.typeToString(checker.getReturnTypeOfSignature(checker.getSignatureFromDeclaration(same)),same,ts.TypeFormatFlags.NoTruncation)+' ';
}
const replacement=annotation+'{ throw new Error("discovery placeholder"); }';
const start=node.body.getStart(source),end=node.body.end;
fs.writeFileSync(file,text.slice(0,start)+replacement+text.slice(end));
console.log(JSON.stringify({file,line:+line,column:+column,name:node.name?.getText(source)||'<anonymous>',start,end,original:text.slice(start,end),replacement}));
