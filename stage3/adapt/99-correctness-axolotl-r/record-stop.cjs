'use strict';
const fs=require('node:fs'),path=require('node:path'),ts=require('typescript'),assert=require('node:assert/strict'),crypto=require('node:crypto');
const file=path.join(process.argv[2],'src/compiler/core.ts'),text=fs.readFileSync(file,'utf8'),source=ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const statement=source.statements.find(n=>ts.isVariableStatement(n)&&n.declarationList.declarations.some(d=>d.name.getText(source)==='elementAt'));assert(statement);
const casts=[];function visit(n){if(ts.isAsExpression(n))casts.push(n);ts.forEachChild(n,visit);}visit(statement);
const cast=casts.find(n=>n.expression.getText(source)==='Array.prototype'&&n.type.kind===ts.SyntaxKind.AnyKeyword);assert(cast);
const pos=source.getLineAndCharacterOfPosition(cast.getStart(source)),owner=statement.getText(source);assert(owner.split(/\r?\n/).length<30);
console.log(JSON.stringify({where:`core.ts:${pos.line+1}:${pos.character+1}`,expression:cast.getText(source),target:'any',enclosingFunction:null,owner:'elementAt',declaration:owner,lines:owner.split(/\r?\n/).length,sha256:crypto.createHash('sha256').update(owner).digest('hex'),refusal:{file:'internal/lower/cast_proof.go',function:'castProof',line:197,reasonConstructionLine:188,conditionLine:196,condition:'source or target has TypeFlagsAny',caller:'castRefusal'}},null,2));
