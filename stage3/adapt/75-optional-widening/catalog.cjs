#!/usr/bin/env node
'use strict';
const fs=require('node:fs');
const path=require('node:path');
const ts=require(process.env.TSC_ADAPT_TYPESCRIPT || 'typescript');
if(ts.version!=='6.0.3') throw new Error('expected TypeScript 6.0.3');
const [tree,input]=process.argv.slice(2);
const config=path.join(tree,'src/compiler/tsconfig.json');
const read=ts.readConfigFile(config,ts.sys.readFile);
if(read.error) throw new Error('config read failed');
const parsed=ts.parseJsonConfigFileContent(read.config,ts.sys,path.dirname(config));
const program=ts.createProgram(parsed.fileNames,parsed.options);
const checker=program.getTypeChecker();
function location(n){const f=n.getSourceFile(),p=f.getLineAndCharacterOfPosition(n.getStart(f));return {file:path.relative(tree,f.fileName),line:p.line+1,column:p.character+1,kind:ts.SyntaxKind[n.kind],name:n.name?.getText(f)};}
function declarations(type){const symbol=type.aliasSymbol || type.getSymbol();return (symbol?.declarations || []).map(location);}
const rows=[];
for(const line of fs.readFileSync(input,'utf8').trim().split('\n')) {
 const r=JSON.parse(line);if(r.summary)continue;
 const file=program.getSourceFile(path.join(tree,'src/compiler',r.File));
 if(!file)throw new Error(`missing ${r.File}`);
 const start=file.getPositionOfLineAndCharacter(r.Line-1,r.Column-1);
 let node;
 function visit(n){if(n.getStart(file)===start && [ts.SyntaxKind[n.kind], 'Kind'+ts.SyntaxKind[n.kind]].includes(r.Kind) && n.getText(file)===r.Text)node=n;if(n.pos<=start && start<n.end)ts.forEachChild(n,visit);}
 visit(file);
 if(!node)throw new Error(`site moved: ${r.File}:${r.Line}:${r.Column}`);
 const value=ts.isAsExpression(node) || ts.isTypeAssertionExpression(node) ? node.expression:node;
 const source=checker.getTypeAtLocation(value);
 const apparent=checker.getApparentType(source);
 const contextual=ts.isAsExpression(node) || ts.isTypeAssertionExpression(node)?checker.getTypeFromTypeNode(node.type):checker.getContextualType(node);
 const property=contextual && checker.getPropertyOfType(contextual,r.Property);
 rows.push({...r,stock_source:checker.typeToString(source),source_declarations:declarations(source),apparent_declarations:declarations(apparent),contextual_target:contextual && checker.typeToString(contextual),target_property_declarations:(property?.declarations || []).map(location)});
}
console.log(JSON.stringify({typescript:ts.version,sites:rows.length,rows},null,2));
