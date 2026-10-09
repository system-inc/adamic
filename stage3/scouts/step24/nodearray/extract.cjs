const fs=require('node:fs'),path=require('node:path'),assert=require('node:assert/strict'),ts=require('typescript');
const root=path.resolve(process.argv[2]),out=__dirname,manifest=[];
function extract(file,name) {
    const full=path.join(root,'src/compiler',file),sf=ts.createSourceFile(full,fs.readFileSync(full,'utf8'),ts.ScriptTarget.Latest,true);
    const matches=[];
    function visit(n){if(ts.isFunctionDeclaration(n)&&n.name?.text===name&&n.body)matches.push(n);ts.forEachChild(n,visit);}
    visit(sf);assert.equal(matches.length,1,`${file}:${name}`);
    const n=matches[0],p=sf.getLineAndCharacterOfPosition(n.getStart(sf));
    return {file:'src/compiler/'+file,line:p.line+1,start:n.getStart(sf),end:n.end,text:n.getText(sf)};
}
const make=extract('factory/nodeFactory.ts','createNodeArray');
const aggregate=extract('factory/nodeFactory.ts','aggregateChildrenFlags');
const test=extract('utilitiesPublic.ts','isNodeArray');
const parserArray=extract('parser.ts','createNodeArray');
const pos=extract('utilities.ts','setTextRangePos'),end=extract('utilities.ts','setTextRangeEnd'),range=extract('utilities.ts','setTextRangePosEnd');
const missing=extract('parser.ts','createMissingList'),isMissing=extract('parser.ts','isMissingList');
const common=`interface Node { readonly kind: number; readonly transformFlags: number; }
interface ReadonlyTextRange { readonly pos: number; readonly end: number; }
interface TextRange { pos: number; end: number; }
// The required-but-possibly-undefined transformFlags slot is represented truthfully here.
interface NodeArray<T extends Node> extends ReadonlyArray<T>, ReadonlyTextRange {
    readonly hasTrailingComma: boolean;
    transformFlags: number | undefined;
}
interface MutableNodeArray<T extends Node> extends Array<T>, TextRange {
    hasTrailingComma: boolean;
    transformFlags: number | undefined;
}
const TransformFlags = { None: 0 };
const emptyArray: readonly Node[] = [];
const Debug = { attachNodeArrayDebugInfo: <T extends Node>(_array: NodeArray<T>): void => {} };
function hasProperty(value: object, key: string): boolean { return Object.hasOwn(value, key); }
// All drivers use identifiers; their child flags are propagated without subtree exclusions.
function propagateChildFlags(child: Node): number { return child.transformFlags; }
${aggregate.text}
${test.text}
function stamp(array: NodeArray<Node>): string {
    return array.length + ':' + array.pos + ':' + array.end + ':' + array.hasTrailingComma + ':' + array.transformFlags;
}
const one: Node = { kind: 80, transformFlags: 1 };
const two: Node = { kind: 80, transformFlags: 2 };
`;
const factory=common+make.text+'\n';
const nestedFactory=common+`function makeFactory() {\n${make.text}\nreturn { createNodeArray };\n}\nconst factoryCreateNodeArray = makeFactory().createNodeArray;\n`;
const rangeCode=[pos,end,range].map(x=>x.text).join('\n');
function save(file,spans,body,mutation,description) {
    const source=spans.map(s=>`// From TypeScript 6.0.3, ${s.file}:${s.line}`).join('\n')+'\n// Step 24: NodeArray fixed layout and field presence.\n'+body;
    fs.writeFileSync(path.join(out,file),source.replace(/\r\n/g,'\n'));
    manifest.push({file,description,spans,mutant:mutation});
}
save('01_factory_clone.a',[make,aggregate,test,pos,end,range],factory+rangeCode+`
const smallInput: readonly Node[] = [one, two];
const small = createNodeArray(smallInput, true);
setTextRangePosEnd(small, 10, 20);
const same = createNodeArray(small);
const changed = createNodeArray(small, false);
console.log(stamp(small));
console.log(stamp(changed));
console.log('small-copy:' + (small !== smallInput) + ' same:' + (same === small) + ' clone:' + (changed !== small) + ' element:' + (changed[0] === one));
const largeInput: readonly Node[] = [one, two, one, two, one];
const large = createNodeArray(largeInput);
console.log('large-reused:' + (large === largeInput) + ' ' + stamp(large));
const empty = createNodeArray<Node>();
console.log(stamp(empty));
`,{find:'array.end = elements.end;',replace:'array.end = -1;'},'Small-array copy, unchanged NodeArray reuse, changed-comma clone preserving range/flags, large-array promotion and empty layout.');
save('02_parser_range.a',[make,aggregate,test,parserArray,pos,end,range],nestedFactory+rangeCode+`
const scanner = { getTokenFullStart: (): number => 42 };
${parserArray.text}
const implicitEnd = createNodeArray([one, two], 7);
const explicitEnd = createNodeArray([one], 9, 11, true);
console.log(stamp(implicitEnd));
console.log(stamp(explicitEnd));
const sliced = explicitEnd.slice();
console.log('slice-pos:' + hasProperty(sliced, 'pos') + ' slice-end:' + hasProperty(sliced, 'end'));
console.log('slice-comma:' + hasProperty(sliced, 'hasTrailingComma') + ' slice-flags:' + hasProperty(sliced, 'transformFlags'));
const rebuilt = createNodeArray(sliced, 12, 15);
console.log(stamp(rebuilt));
`,{find:'end ?? scanner.getTokenFullStart()',replace:'end ?? 0'},'Parser range initialization with implicit and explicit ends; slice drops metadata, then reconstruction restores the layout.');
save('03_repair_flags.a',[make,aggregate,test],factory+`
const array = createNodeArray([one, two]);
array.transformFlags = undefined;
console.log('before-own:' + hasProperty(array, 'transformFlags') + ' before-undefined:' + (array.transformFlags === undefined));
const repaired = createNodeArray(array);
console.log('same:' + (repaired === array) + ' after-own:' + hasProperty(repaired, 'transformFlags') + ' flags:' + repaired.transformFlags);
`,{find:'aggregateChildrenFlags(elements as MutableNodeArray<T>);',replace:'// mutant: omitted aggregation'},'The factory repairs a present-undefined transformFlags slot when reusing an existing NodeArray. Driver supplies the possibly-undefined state mentioned by upstream.');
save('04_missing_presence.a',[make,aggregate,test,parserArray,pos,end,range,missing,isMissing],nestedFactory+rangeCode+`
const scanner = { getTokenFullStart: (): number => 42 };
function getNodePos(): number { return 17; }
${parserArray.text}
interface MissingList<T extends Node> extends NodeArray<T> { isMissingList: true; }
${missing.text}
${isMissing.text}
const ordinary = createNodeArray<Node>([], 17);
const missing = createMissingList<Node>();
console.log('ordinary-own:' + hasProperty(ordinary, 'isMissingList') + ' missing:' + isMissingList(ordinary));
console.log('missing-own:' + hasProperty(missing, 'isMissingList') + ' missing:' + isMissingList(missing));
// A presence probe, not a claim that tsc itself writes undefined to isMissingList.
const presentUndefined = createNodeArray<Node>([], 17) as NodeArray<Node> & { isMissingList?: true | undefined };
presentUndefined.isMissingList = undefined;
console.log('undefined-own:' + hasProperty(presentUndefined, 'isMissingList') + ' missing:' + isMissingList(presentUndefined));
`,{find:'list.isMissingList = true;',replace:'list.isMissingList = undefined;'},'MissingList adds a field outside the base declaration; ordinary lists have it absent. A marked driver probe holds present-undefined distinct from absent.');
fs.writeFileSync(path.join(out,'manifest.json'),JSON.stringify(manifest,null,2)+'\n');
console.log('extracted four fixtures with unchanged upstream function bodies');
