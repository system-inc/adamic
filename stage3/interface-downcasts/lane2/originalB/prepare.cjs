// Full declarations and exact original read spans, from independent upstream.
const fs = require('node:fs');
const path = require('node:path');
const cp = require('node:child_process');
const crypto = require('node:crypto');
const ts = require('../../../api/node_modules/typescript');
const [root, out] = process.argv.slice(2).map(p => path.resolve(p));
if (!root || !out) throw Error('usage: prepare.cjs <pristine-upstream> <declarations>');
cp.execFileSync(process.execPath, [path.resolve(__dirname, '../../lane4b/original/prepare.cjs'), root, out], {stdio: 'inherit'});
const manifest = JSON.parse(fs.readFileSync(path.join(out, 'original-manifest.json')));
const queue = JSON.parse(fs.readFileSync(path.resolve(__dirname, '../lazy-array-priority.json'))).candidate_queue;
const targets = JSON.parse(fs.readFileSync(path.join(__dirname, 'config.json'))).targets;
const program = ts.createProgram(['types.ts', 'watchUtilities.ts', 'builder.ts'].map(f => path.join(root, 'src/compiler', f)), {target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext, moduleResolution: ts.ModuleResolutionKind.Bundler, strict: true, types: []});
const checker = program.getTypeChecker();
const privateSource = program.getSourceFile(path.join(root, 'src/compiler/watchUtilities.ts'));
const privateNames = ['Canonicalized', 'SortedAndCanonicalizedMutableFileSystemEntries'];
const printer = ts.createPrinter();
let privateText = "import type { SortedArray } from './compiler/corePublic.d.ts';\n";
for (const name of privateNames) {
 const declaration = privateSource.statements.find(n => n.name?.text === name);
 if (!declaration) throw Error('missing private original declaration ' + name);
 privateText += 'export ' + printer.printNode(ts.EmitHint.Unspecified, declaration, privateSource) + '\n';
}
fs.writeFileSync(path.join(out, 'private-watch.d.ts'), privateText);
const hash = file => crypto.createHash('sha256').update(fs.readFileSync(file)).digest('hex');
manifest.declarations['private-watch.d.ts'] = hash(path.join(out, 'private-watch.d.ts'));
manifest.private_source_sha256 = hash(privateSource.fileName);
manifest.fields = {};
manifest.pairs = [];
const probes = [];
const addFields = type => {const name = checker.typeToString(type); manifest.fields[name] = checker.getPropertiesOfType(type).map(p => p.name).sort(); return name;};
for (const [target, field] of targets) {
 const pair = queue.find(p => p.type === target && p.field === field);
 if (!pair || pair.read_count !== 1 || pair.sites.length !== 1) throw Error('rank drift: ' + target + '.' + field);
 const site = pair.sites[0];
 const source = program.getSourceFile(path.join(root, site.file));
 if (!source || source.text.slice(site.start, site.end) !== site.text) throw Error('original witness drift');
 let access;
 const visit = node => {if (ts.isPropertyAccessExpression(node) && node.getStart(source) === site.start && node.end === site.end) access = node; ts.forEachChild(node, visit);};
 visit(source);
 if (!access) throw Error('original access absent');
 const receiver = checker.getNonNullableType(checker.getTypeAtLocation(access.expression));
 const member = checker.getPropertyOfType(receiver, field);
 const declared = checker.getTypeOfSymbolAtLocation(member, member.valueDeclaration || member.declarations[0]);
 if (checker.typeToString(declared) !== pair.declared_type) throw Error('original declared type drift: ' + target);
 const array = checker.getNonNullableType(declared);
 let element = checker.getIndexTypeOfType(array, ts.IndexKind.Number);
 if (element?.isUnion()) element = element.types.find(t => checker.getPropertyOfType(t, 'pos')) || element.types[0];
 if (!element) throw Error('not an array field');
 const receiverNames = receiver.isUnion() ? receiver.types.map(addFields) : [addFields(receiver)];
 const nested = target === 'IncrementalMultiFileEmitBuildInfo';
 const stringElement = !!(element.flags & ts.TypeFlags.String);
 const scalarField = stringElement ? undefined : checker.getPropertiesOfType(element).find(p => ['pos', 'flags'].includes(p.name)) || checker.getPropertiesOfType(element).find(p => p.name === 'path');
 if (!stringElement && !scalarField && !nested) throw Error('no represented original element field: ' + checker.typeToString(element));
 const scalarType = scalarField && checker.getTypeOfSymbolAtLocation(scalarField, scalarField.valueDeclaration || scalarField.declarations[0]);
 const stringField = scalarType && !!(scalarType.flags & ts.TypeFlags.String);
 const elementName = stringElement || nested ? undefined : addFields(element);
 const kind = checker.getPropertyOfType(receiver, 'kind');
 const kindType = kind && checker.getTypeOfSymbolAtLocation(kind, kind.valueDeclaration || kind.declarations[0]);
 const kindValue = kindType && (kindType.value ?? (kindType.isUnion() ? kindType.types.find(t => typeof t.value === 'number')?.value : undefined));
 const elementKind = !stringElement && checker.getPropertyOfType(element, 'kind');
 const elementTag = elementKind && checker.getTypeOfSymbolAtLocation(elementKind, elementKind.valueDeclaration || elementKind.declarations[0]).value;
 let baseField = 'flags';
 let baseType = 'number';
 let baseValue = '8';
 if (kind) { baseField = 'kind'; baseValue = String(typeof kindValue === 'number' ? kindValue : 1); }
 else if (nested) { baseField = 'version'; baseType = 'string'; baseValue = "'1'"; }
 else if (target.startsWith('ParsedCommandLine')) { baseField = 'compileOnSave'; baseType = 'boolean'; baseValue = 'false'; }
 else if (target === 'SortedAndCanonicalizedMutableFileSystemEntries') { baseField = 'files'; baseType = 'string[]'; baseValue = "['kept']"; }
 else if (target === 'WatchOptions') { baseField = 'watchFile'; baseValue = '0'; }
 else if (target === 'ConditionalRoot') { baseField = 'isDistributive'; baseType = 'boolean'; baseValue = 'false'; }
 else if (target.startsWith('ConfigFileSpecs')) { baseField = 'isDefaultIncludeSpec'; baseType = 'boolean'; baseValue = 'false'; }
 const tag = `${baseField}: ${baseValue}, `;
 const elementPrefix = typeof elementTag === 'number' ? `kind: ${elementTag}, ` : '';
 const leaf = scalarField?.name;
 const goodElement = nested ? '[7]' : stringElement ? "'a'" : `{${elementPrefix}${leaf}: ${stringField ? "'a'" : 7}}`;
 const badElement = nested ? "['bad']" : stringElement ? '42' : `{${elementPrefix}${leaf}: ${stringField ? 42 : "'bad'"}}`;
 const imports = [...new Set(target.match(/[A-Za-z_$][\w$]*/g).filter(n => n !== 'undefined'))];
 const module = nested ? 'original-tsc-builder' : target === 'SortedAndCanonicalizedMutableFileSystemEntries' ? 'original-tsc-private' : 'original-tsc-types';
 const prefix = target.replace(/[^A-Za-z0-9]/g, '-').toLowerCase() + '-' + field.toLowerCase();
 const modes = ['good', 'lazy', 'lazy-element', 'bad', 'alias'];
 if ((member.flags & ts.SymbolFlags.Optional) || declared.isUnion() && declared.types.some(t => t.flags & ts.TypeFlags.Undefined)) modes.push('absent', 'undefined');
 if (target.includes('undefined')) modes.push('receiver-undefined');
 for (const mode of modes) {
  let payload = `[${mode === 'bad' ? badElement : goodElement}]`;
  if (mode === 'lazy') payload = '7';
  if (mode === 'lazy-element') payload = `[${goodElement}, ${badElement}]`;
  if (mode === 'undefined') payload = 'undefined';
  const raw = `{${tag}${mode === 'absent' ? '' : field + ': ' + payload}}`;
  let output = stringElement || stringField ? 'a\n' : '7\n';
  let demand = nested ? 'console.log(`${items[0]![0]!}`);' : stringElement ? 'console.log(items.slice(0, 1).join(\';\'));' : 'console.log(`${items[0]!.' + leaf + '}`);';
  if (mode === 'bad') output = stringElement || stringField ? '42\n' : 'bad\n';
  if (mode === 'lazy') {demand = "console.log('kept');"; output = 'kept\n';}
  if (['absent', 'undefined', 'receiver-undefined'].includes(mode)) output = 'absent\n';
  let body = `import type { ${imports.join(', ')} } from '${module}';\ninterface Base { readonly ${baseField}${['ParsedCommandLine | undefined', 'WatchOptions'].includes(target) ? '?' : ''}: ${baseType}; }\nfunction read(base: Base${target.includes('undefined') ? ' | undefined' : ''}): void {\n const viewed = ${target.includes('undefined') ? 'base === undefined ? undefined : base as ' + target.replace(' | undefined', '') : 'base as ' + target};\n`;
  if (mode !== 'lazy') body += ` const items = viewed${target.includes('undefined') ? '?' : ''}.${field};\n if (items === undefined) { console.log('absent'); return; }\n`;
  if (mode === 'alias') {
   body += ' mutate();\n';
   output = stringElement || stringField ? 'second\n' : '9\n';
  }
  body += ' ' + demand + '\n}\n';
  body += `const raw = ${raw};\n`;
  if (mode === 'alias' && nested) body += `function mutate(): void { raw.${field}[0]![0] = 9; }\n`;
  else if (mode === 'alias') body += `function mutate(): void { raw.${field}[0] = ${nested ? '[9]' : stringElement ? "'second'" : `{${elementPrefix}${leaf}: ${stringField ? "'second'" : 9}}`}; }\n`;
  body += `read(${mode === 'receiver-undefined' ? 'undefined' : 'raw'});\n`;
  const name = prefix + '-' + mode;
  fs.writeFileSync(path.join(__dirname, name + '.a'), body);
  probes.push({name, source: output, refusal: target === 'ParsedCommandLine | undefined' && field === 'projectReferences' && mode !== 'lazy' ? 'Adamic 0.1 refuses checked view read of field path with unsupported intersection contract; prove or implement the intersection contract before reading this field' : '', diagnostic: mode === 'absent' && !(member.flags & ts.SymbolFlags.Optional) ? `field read failed: viewed${target.includes('undefined') ? '?' : ''}.${field} is not initialized; expected ${pair.declared_type}, found missing` : mode !== 'bad' ? '' : nested ? 'element read failed: items[0]![0] expected IncrementalBuildInfoFileId, found string' : stringElement ? 'element read failed: items.slice(0, 1)[element] expected string, found number' : `field read failed: items[0]!.${leaf} is not a ${checker.typeToString(scalarType)}; expected ${checker.typeToString(scalarType)}, found ${stringField ? 'number' : 'string'}`, contracts: mode === 'lazy' ? receiverNames : [...receiverNames, ...(elementName ? [elementName] : [])], pair: target + '.' + field, mode, stringElement, nested, leaf});
 }
 manifest.pairs.push(pair);
}
probes.push({name: 'incremental-multi-file-ids-replacement', source: '9\n', diagnostic: 'element read failed: <array write> expected array, found uncertified source element contract', contracts: ['IncrementalMultiFileEmitBuildInfo'], pair: 'IncrementalMultiFileEmitBuildInfo.fileIdsList', mode: 'replacement', codeRequired: true});
fs.writeFileSync(path.join(out, 'arrayB-manifest.json'), JSON.stringify(manifest, null, 2) + '\n');
fs.writeFileSync(path.join(__dirname, 'probes.json'), JSON.stringify(probes, null, 2) + '\n');
console.log(`Prepared ${targets.length} one-read pairs and ${probes.length} probes from complete originals`);
