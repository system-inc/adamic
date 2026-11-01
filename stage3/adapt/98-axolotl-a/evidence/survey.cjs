const ts = require("typescript"), fs = require("fs"), path = require("path");
const want = {"transformers/es2017.ts": ["node.declarations[0]", "node.parameters[i]", "outerParameters[i]"],
 "transformers/classFields.ts": ["privateInstanceMethodsAndAccessors[0]", "superPath[superPathDepth]", "statementsIn[superStatementIndex]", "statementsIn[statementOffset]", "constructor.body.statements[statementOffset]"]};
for (const [file, exprs] of Object.entries(want)) {
  const text = fs.readFileSync(path.join(process.argv[2], "src/compiler", file), "utf8");
  const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
  const found = [];
  (function v(n){ if (ts.isElementAccessExpression(n) && exprs.includes(n.getText(sf))) found.push(n); ts.forEachChild(n, v); })(sf);
  for (const n of found) { let f=n.parent; while(f && !(ts.isFunctionDeclaration(f)&&f.name)) f=f.parent;
    const p = sf.getLineAndCharacterOfPosition(n.getStart(sf));
    console.log(JSON.stringify({file, line:p.line+1, column:p.character+1, function:f&&f.name.text, expression:n.getText(sf), parent: ts.SyntaxKind[n.parent.kind]})); }
}
