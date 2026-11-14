#!/usr/bin/env node
// Stopped codemod prototype: symbol-resolved preflight, no source writes.
// The emission phase is intentionally unimplemented after the real input failed.
const ts = require('../../api/node_modules/typescript');
const fs = require('fs'), path = require('path');
const [rootArg, inventory, output] = process.argv.slice(2);
if (!output) throw Error('usage: node codemod.cjs adapted-root heat.json result.json');
const root = path.resolve(rootArg);
const cfg = ts.getParsedCommandLineOfConfigFile(path.join(root, 'src/compiler/tsconfig.json'), {}, {...ts.sys, onUnRecoverableConfigFileDiagnostic: d => { throw Error(ts.flattenDiagnosticMessageText(d.messageText, '\n')); }});
const roots = ts.sys.readDirectory(path.join(root, 'src/compiler'), ['.ts']);
const program = ts.createProgram(roots, cfg.options), checker = program.getTypeChecker();
if (ts.version !== '6.0.3') throw Error('expected TypeScript 6.0.3');
const sites = JSON.parse(fs.readFileSync(inventory)).sites.filter(s => s.public);
if (sites.length !== 19) throw Error('expected 19 public sites');
const failures = [], results = [];
function walk(n, fn) { fn(n); ts.forEachChild(n, c => walk(c, fn)); }
function location(n) { const f = n.getSourceFile(), p = f.getLineAndCharacterOfPosition(n.getStart()); return `${path.relative(root,f.fileName)}:${p.line+1}:${p.character+1}`; }
for (const site of sites) {
 const sf = program.getSourceFile(path.join(root,site.file));
 let binding;
 walk(sf,n=>{if(ts.isObjectBindingPattern(n) && sf.getLineAndCharacterOfPosition(n.getStart()).line+1===site.line) binding=n;});
 if(!binding) throw Error(`inventory mismatch ${site.file}:${site.line}`);
 const declaration = binding.parent, receiver = declaration.initializer;
 if(!receiver || !ts.isIdentifier(receiver)) throw Error(`unsupported receiver ${location(binding)}`);
 const row={file:site.file,site:location(binding),receiver:receiver.text,members:[]}; results.push(row);
 let scope=declaration;
 while(scope.parent && !ts.isFunctionLike(scope) && !ts.isModuleBlock(scope)) scope=scope.parent;

 for(const element of binding.elements) {
  const member=(element.propertyName||element.name).getText(sf), meta=site.members.find(m=>m.name===member);
  const property=checker.getTypeAtLocation(receiver).getProperty(member);
  const method=!!property?.declarations?.some(ts.isMethodSignature);
  if (!meta || method !== meta.method) throw Error(`method inventory mismatch ${location(element)}`);
  // Existing function properties and the factory property retain their bindings.
  // Defaulted method locals are the proposed readonly-view exception.
  if(!meta?.method || element.initializer) { continue; }
  if(!ts.isIdentifier(element.name)) {failures.push({where:location(element),reason:'nested binding'});continue;}
  const symbol=checker.getSymbolAtLocation(element.name), refs=[];
  if (!symbol) throw Error(`unresolved binding ${location(element)}`);
  walk(sf,n=>{if(ts.isIdentifier(n)&&n!==element.name&&checker.getSymbolAtLocation(n)===symbol) refs.push(n);});
  const m={member,local:element.name.text,directCalls:0,unsafe:[]};row.members.push(m);
  for(const n of refs) {
   const p=n.parent;
   if(ts.isCallExpression(p)&&p.expression===n&&!p.questionDotToken) {
    m.directCalls++;
   } else { const f={where:location(n),member,reason:'local used other than as a direct callee',syntax:ts.SyntaxKind[p.kind],text:p.getText(sf)};m.unsafe.push(f);failures.push(f); }
  }
  // A receiver shadow at a use would change binding on rewriting.
  const receiverSymbol=checker.getSymbolAtLocation(receiver);
  for(const n of refs) {
   const visible=checker.resolveName(receiver.text,n,ts.SymbolFlags.Value,false);
   if(visible!==receiverSymbol) failures.push({where:location(n),member,reason:'receiver is shadowed'});
  }
  walk(scope,n=>{if(ts.isIdentifier(n)&&n.text===element.name.text&&n!==element.name&&checker.getSymbolAtLocation(n)!==symbol && ts.isDeclarationName(n)) failures.push({where:location(n),member,reason:'shadowed local declaration'});});
 }

}
const report={roots:roots.length,sites:results.length,status:failures.length?'stopped':'preflight-passed',rewrittenUses:0,linesChanged:0,failures,files:results};
// Deliberately never emit a partial compiler. Unsafe findings require a decision.
if(!failures.length) throw Error('preflight passed: default-view construction must be implemented before source emission');
fs.writeFileSync(output,JSON.stringify(report,null,2)+'\n');
for(const f of failures) console.error(`${f.where}: ${f.reason}${f.member?' ('+f.member+')':''}`);
process.exitCode=2;
