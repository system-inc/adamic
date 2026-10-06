// Recorded Node observations; CI consumes these without Node or npm.
const fs = require('node:fs');
const zlib = require('node:zlib');
const crypto = require('node:crypto');
const vm = require('node:vm');
const {RegExpParser, visitRegExpAST} = require('@eslint-community/regexpp');
const inventory = JSON.parse(fs.readFileSync(`${__dirname}/inventory.json`, 'utf8'));
if (!process.version.startsWith('v24.')) throw Error('Node 24 required');
const seed = 0xc0ae2026;
const units = s => Array.from({length:s.length}, (_,i)=>s.charCodeAt(i));
const fromUnits = a => a.map(c=>String.fromCharCode(c)).join('');
const observations = {version:1, node:process.version, v8:process.versions.v8, unicode:process.versions.unicode, seed, algorithm:'xorshift32, seed XOR (pattern id + 1); 24 bounded AST witnesses, 32 alphabet strings of 0..16 code points; boundary probes; complete g/y exec walks with explicit AdvanceStringIndex after empty matches', inventorySHA256:crypto.createHash('sha256').update(fs.readFileSync(`${__dirname}/inventory.json`)).digest('hex'), patterns:[], cases:[]};
const parser = new RegExpParser({ecmaVersion:2025});
const universe = Array.from({length:128}, (_,i)=>i).concat([0x85,0xa0,0xe9,0x17f,0x3b1,0x2003,0x2028,0x2029,0x212a,0x4e16,0xfeff,0xd800,0xdc00,0x1f30d]);
let sequence = 0;
for (const p of inventory.patterns) {
  let randomState = (seed ^ (p.id + 1)) >>> 0;
  function random(n) { randomState ^= randomState << 13; randomState ^= randomState >>> 17; randomState ^= randomState << 5; return (randomState >>> 0) % n; }
  const info = {id:p.id, alphabet:[], nodeError:'', nodePatternError:'', sourceError:'', generatorError:'', inputCount:0, caseCount:0};
  observations.patterns.push(info);
  let regex;
  try { if (p.literalToken !== undefined) new vm.Script(`(${p.literalToken})`); } catch (e) { info.sourceError=e.message; }
  try { regex = new RegExp(fromUnits(p.patternUnits), p.flags.includes('d')?p.flags:p.flags+'d'); } catch (e) { info.nodePatternError=e.message; }
  info.nodeError=info.nodePatternError || info.sourceError;
  if (info.nodeError) continue;
  const inputs = new Map();
  function add(s, provenance) {
    const key = JSON.stringify(units(s));
    if (!inputs.has(key)) inputs.set(key, {input:units(s), sources:[]});
    inputs.get(key).sources.push(provenance);
  }
  for (const fixture of p.inputs) add(fromUnits(fixture.units), fixture);
  let ast;
  const alphabet = new Set(), atoms = new Map();
  function choices(node) {
    if (atoms.has(node)) return atoms.get(node);
    let result=[];
    if (node.type === 'Character') result = [String.fromCodePoint(node.value)];
    else if (node.type === 'CharacterClassRange') result = [node.min.value, node.max.value, Math.floor((node.min.value+node.max.value)/2)].map(c=>String.fromCodePoint(c));
    else {
      try {
        const flags = p.flags.replace(/[dgym]/g, '');
        const probe = new RegExp(`^(?:${node.raw})$`,flags);
        result = universe.map(c=>String.fromCodePoint(c)).filter(s=>probe.test(s));
      } catch (_) { /* Recorded syntax-parser gaps do not change Node's result. */ }
    }
    for (const s of result) alphabet.add(s);
    atoms.set(node,result);
    return result;
  }
  try {
    ast = parser.parsePattern(fromUnits(p.patternUnits),0,p.patternUnits.length,{unicode:p.flags.includes('u'),unicodeSets:p.flags.includes('v')});
    visitRegExpAST(ast, {
      onCharacterEnter: choices, onCharacterClassRangeEnter: choices, onCharacterSetEnter: choices,
      onCharacterClassEnter: choices, onExpressionCharacterClassEnter: choices,
    });
  } catch (e) { info.generatorError=e.message; }
  // If a parser cannot represent new syntax, derive literal code points from its
  // raw spelling and label the gap. This is a finite alphabet sample, not a proof.
  if (!ast) for (const s of fromUnits(p.patternUnits)) alphabet.add(s);
  info.alphabet = Array.from(alphabet).sort((a,b)=>a.codePointAt(0)-b.codePointAt(0)).map(units);
  const letters = info.alphabet.map(fromUnits);
  const captures = new Map();
  function witness(node, depth=0) {
    if (depth>32) return '';
    switch(node.type) {
      case 'Pattern': case 'Group': case 'CapturingGroup': {
        const value = witness(node.alternatives[random(node.alternatives.length)],depth+1);
        if (node.type==='CapturingGroup') captures.set(node,value);
        return value;
      }
      case 'Alternative': return node.elements.map(n=>witness(n,depth+1)).join('');
      case 'Quantifier': {
        const low=Math.min(node.min,8), high=Math.min(node.max,Math.max(low,3));
        return Array.from({length:low+random(high-low+1)},()=>witness(node.element,depth+1)).join('');
      }
      case 'Backreference': return captures.get(Array.isArray(node.resolved)?node.resolved[0]:node.resolved)||'';
      case 'Assertion': return node.alternatives && !node.negate ? witness(node.alternatives[random(node.alternatives.length)],depth+1) : '';
      case 'StringAlternative': return node.elements.map(n=>witness(n,depth+1)).join('');
      case 'ClassStringDisjunction': return witness(node.alternatives[random(node.alternatives.length)],depth+1);
      default: {const candidates=choices(node); return candidates.length?candidates[random(candidates.length)]:'';}
    }
  }
  if (ast) for(let i=0;i<24;i++) {
    captures.clear(); const s=witness(ast).slice(0,128);
    add(s,{kind:'generated witness',ordinal:i});
    if (i<4) add(s+s,{kind:'repeated witness',ordinal:i});
  }
  for(let i=0;i<32;i++) {
    let s=''; const length=random(17);
    for(let j=0;j<length && letters.length;j++) s+=letters[random(letters.length)];
    add(s,{kind:'generated alphabet',ordinal:i});
  }
  for(const s of ['', 'a', '0', ' ', '\n', '\r\n', 'é', 'Kſ', '🌍', '\ud800a\udc00']) add(s,{kind:'boundary probe'});
  // Single representatives also make otherwise rare named alternatives observable.
  for (const letter of letters.slice(0,128)) add(letter,{kind:'alphabet singleton'});
  info.inputCount=inputs.size;
  for (const entry of inputs.values()) {
    const input=fromUnits(entry.input), stateful=/[gy]/.test(p.flags);
    const starts=stateful?Array.from(new Set([0,1,input.length,input.length+1])):[0, input.length+1];
    for (const start of starts) {
      regex.lastIndex=start;
      const series=sequence++;
      for(let step=0;step<=input.length+2;step++) {
        const lastIndex=regex.lastIndex, match=regex.exec(input);
        const expected={index:match?match.index:null, captures:match?Array.from(match.indices,x=>x??null):null, values:match?Array.from(match,x=>x===undefined?null:units(x)):null, groups:match?.indices.groups?Object.fromEntries(Object.entries(match.indices.groups).map(([k,v])=>[k,v??null])):null, groupValues:match?.groups?Object.fromEntries(Object.entries(match.groups).map(([k,v])=>[k,v===undefined?null:units(v)])):null, lastIndex:regex.lastIndex};
        observations.cases.push({pattern:p.id, input:entry.input, sources:entry.sources, series, step, lastIndex, expected});
        info.caseCount++;
        if(!stateful || !match) break;
        if(match[0].length===0) {
          let next=regex.lastIndex+1;
          const c=input.charCodeAt(regex.lastIndex), d=input.charCodeAt(regex.lastIndex+1);
          if(/[uv]/.test(p.flags) && c>=0xd800 && c<=0xdbff && d>=0xdc00 && d<=0xdfff) next++;
          regex.lastIndex=next;
        }
        if(step===input.length+2) throw Error('nonterminating exec sequence');
      }
    }
  }
}
fs.writeFileSync(`${__dirname}/observations.json.gz`,zlib.gzipSync(JSON.stringify(observations)+'\n',{level:9}));
console.log(JSON.stringify({node:observations.node,patterns:observations.patterns.length,nodeRejections:observations.patterns.filter(p=>p.nodeError).length,generatorGaps:observations.patterns.filter(p=>p.generatorError).length,inputs:observations.patterns.reduce((n,p)=>n+p.inputCount,0),executions:observations.cases.length}));
