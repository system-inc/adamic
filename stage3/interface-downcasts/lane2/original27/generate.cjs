// Derive fixture receiver tags and complete field lists from original declarations.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('../../../api/node_modules/typescript');
const root = path.resolve(process.argv[2]);
const file = path.join(root, 'src/compiler/types.ts');
const program = ts.createProgram([file], {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, strict: true, types: []});
const checker = program.getTypeChecker();
const symbols = checker.getExportsOfModule(checker.getSymbolAtLocation(program.getSourceFile(file)));
const tagOf = name => {const type = checker.getDeclaredTypeOfSymbol(symbols.find(s => s.name === name)); const prop = checker.getPropertyOfType(type, 'kind'); return checker.getTypeOfSymbolAtLocation(prop, prop.valueDeclaration || prop.declarations[0]).value;};
const targets = [['FunctionLikeDeclaration', 'typeParameters', 3, 'function-typeParameters', 'TypeParameterDeclaration', tagOf('TypeParameterDeclaration')], ['CallExpression | NewExpression', 'typeArguments', 3, 'call-typeArguments', 'TypeNode', tagOf('TypeReferenceNode')], ['JsxElement | JsxFragment', 'children', 3, 'jsx-children', 'JsxText', tagOf('JsxText')]];
const fields = new Set(['TypeParameterDeclaration', 'TypeNode', 'JsxText']);
const jsxArms = checker.getDeclaredTypeOfSymbol(symbols.find(s => s.name === 'JsxChild')).types.map(t => t.symbol.name);
for (const name of jsxArms) fields.add(name);
const probes = [];
for (const [target, field, reads, prefix, defaultElement, defaultKind] of targets) {
 const types = target.split(' | ').flatMap(name => {const type = checker.getDeclaredTypeOfSymbol(symbols.find(s => s.name === name)); return type.isUnion() ? type.types : [type];});
 const arms = types.map(type => {const kind = checker.getPropertyOfType(type, 'kind'); const tag = checker.getTypeOfSymbolAtLocation(kind, kind.valueDeclaration || kind.declarations[0]).value; if (typeof tag !== 'number') throw Error('original tag missing'); fields.add(type.symbol.name); return {type, name: type.symbol.name, tag};});
 const receiverOptional = arms.some(arm => (checker.getPropertyOfType(arm.type, field).flags & ts.SymbolFlags.Optional) !== 0);
 for (const [index, arm] of arms.entries()) {
  const element = arm.name === 'JSDocSignature' ? 'JSDocParameterTag' : defaultElement;
  const elementKind = arm.name === 'JSDocSignature' ? tagOf('JSDocParameterTag') : defaultKind;
  const member = checker.getPropertyOfType(arm.type, field);
  const optional = (member.flags & ts.SymbolFlags.Optional) !== 0;
  const modes = index === 0 || arm.name === 'JSDocSignature' ? ['good', 'lazy', 'wrong-array', 'wrong-pos', 'missing-pos', 'lazy-element', 'wrong-kind', 'map', 'map-bad'] : ['good'];
  if (optional) modes.push('absent', 'undefined');
  for (const mode of modes) {
   let payload = `[{kind: ${elementKind}, pos: 7}]`;
   let read = 'console.log(`${items.length}:${items[0]!.pos}`);';
   let source = '1:7\n';
   let diagnostic = '';
   const contracts = [arm.name];
   if (prefix === 'jsx-children') contracts.push(...jsxArms);
   if (mode === 'lazy') {payload = '7'; read = 'console.log(`${viewed.kind}`);'; source = arm.tag + '\n';}
   if (mode === 'wrong-array') {payload = '7'; read = `console.log(typeof viewed.${field});`; source = 'number\n'; diagnostic = `field read failed: viewed.${field} ${!optional ? 'is not a ' + (arm.name === 'JSDocSignature' ? 'ReadonlyArray<JSDocParameterTag>' : 'NodeArray<' + (prefix === 'jsx-children' ? 'JsxChild' : element) + '>') + '; expected ' + (arm.name === 'JSDocSignature' ? 'ReadonlyArray<JSDocParameterTag>' : 'NodeArray<' + (prefix === 'jsx-children' ? 'JsxChild' : element) + '>') : 'matches no member of NodeArray<' + element + '> | undefined; expected NodeArray<' + element + '> | undefined'}, found number`;}
   if (mode === 'wrong-pos' || mode === 'map-bad') {payload = `[{kind: ${elementKind}, pos: 'bad'}]`; read = 'console.log(`${items[0]!.pos}`);'; source = 'bad\n'; diagnostic = 'field read failed: items[0]!.pos is not a number; expected number, found string';}
   if (mode === 'missing-pos') {payload = `[{kind: ${elementKind}}]`; read = 'console.log(`${items[0]!.pos}`);'; source = 'undefined\n'; diagnostic = 'field read failed: items[0]!.pos is not initialized; expected number, found missing';}
   if (mode === 'lazy-element') {payload = `[{kind: ${elementKind}, pos: 7}, {kind: ${elementKind}, pos: undefined}]`; source = '2:7\n';}
   if (mode === 'wrong-kind') {payload = '[42]'; read = 'console.log(typeof items[0]);'; source = 'number\n'; diagnostic = `element read failed: items[0] expected ${prefix === 'jsx-children' ? 'JsxChild' : element}, found number`;}
   if (mode === 'map' || mode === 'map-bad') {read = "console.log(items.map(item => `${item.pos}`).join(';'));"; source = mode === 'map' ? '7\n' : 'bad\n'; if (mode === 'map-bad') diagnostic = 'field read failed: item.pos is not a number; expected number, found string';}
   if (mode === 'absent' || mode === 'undefined') {payload = mode === 'undefined' ? 'undefined' : undefined; read = "console.log(items === undefined ? 'absent' : 'present');"; source = 'absent\n';}
   if (!['lazy', 'wrong-array', 'absent', 'undefined'].includes(mode)) {
    if (element === 'ModifierLike') {fields.add('ExportKeyword'); contracts.push('ExportKeyword');} else {fields.add(element); contracts.push(element);}
   }
   let body = `import type { ${[...new Set(target.split(' | '))].join(', ')} } from 'original-tsc-types';\ninterface Base { readonly kind: number; }\nfunction read(base: Base): void {\n const viewed = base as ${target};\n`;
   if (mode !== 'lazy' && mode !== 'wrong-array') body += ` const items = viewed.${field};\n`;
   if (mode !== 'lazy' && receiverOptional && !['wrong-array', 'absent', 'undefined'].includes(mode)) body += " if (items === undefined) return;\n";
   body += ` ${read}\n}\nconst raw = { kind: ${arm.tag}${payload === undefined ? '' : ', ' + field + ': ' + payload} };\nread(raw);\n`;
   const name = `${prefix}-${arm.tag}-${mode}`;
   fs.writeFileSync(path.join(__dirname, name + '.a'), body);
   probes.push({name, source, diagnostic, contracts});
  }
 }
}
fs.writeFileSync(path.join(__dirname, 'config.json'), JSON.stringify({targets: targets.map(t => t.slice(0, 3)), fields: [...fields].sort()}, null, 2) + '\n');
fs.writeFileSync(path.join(__dirname, 'probes.json'), JSON.stringify(probes, null, 2) + '\n');
console.log('Generated ' + probes.length + ' original union array probes');
