const ts = require("/home/user/adamic/stage3/api/node_modules/typescript");
const fs = require("fs");
const text = fs.readFileSync(process.argv[2], "utf8");
const sf = ts.createSourceFile("es2015.ts", text, ts.ScriptTarget.Latest, true);
const want = ["node.declarations[0]", "declarations[0]", "funcStatements[classBodyStart]", "segments[0]"];
const groups = {};
(function v(n){ if (ts.isElementAccessExpression(n)) { const t=n.getText(sf); if (want.includes(t)) (groups[t] ??= []).push(n);} ts.forEachChild(n,v);})(sf);
for (const t of want) groups[t].forEach((n,i)=>{
  let f=n.parent; while(f && !ts.isFunctionLike(f)) f=f.parent;
  const lc=sf.getLineAndCharacterOfPosition(n.getStart(sf));
  const outer = ts.isNonNullExpression(n.parent)? n.parent : n;
  console.log(t, `${i+1}/${groups[t].length}`, `${lc.line+1}:${lc.character+1}`, ts.SyntaxKind[f.kind], f.name?.text, "asserted="+ts.isNonNullExpression(n.parent), ts.SyntaxKind[outer.parent.kind]);
});
