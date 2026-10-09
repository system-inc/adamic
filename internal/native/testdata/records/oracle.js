// Node is the independent oracle for successful own operations. The record-only missing
// prototype-member stops are checked separately against their ruled diagnostic and Node's member list.
const mode = process.argv[2];
function snapshot(r) {
  console.log(JSON.stringify(Object.keys(r)));
  console.log(JSON.stringify(Object.values(r)));
  console.log(JSON.stringify(r));
}
function pairs(r) {
  for (const k in r) { console.log(JSON.stringify(k)); console.log(r[k]); }
}
function define(r, k, v) {
  Object.defineProperty(r, k, {value:v, writable:true, enumerable:true, configurable:true});
}
if (mode === 'semantics') {
  const r = {};
  snapshot(r);
  const names = ['10','2','-1','01','4294967295','1.5','-0','0','4294967294','1e0','','toString','constructor','__proto__','hasOwnProperty','世界🌍','é'];
  names.forEach((k, i) => define(r, k, i));
  define(r, '9'.repeat(4096), 99);
  snapshot(r);
  r['2'] = 100;
  console.log(+delete r['01'], +delete r['10'], +delete r.absent);
  r['01'] = 101; r['10'] = 102;
  snapshot(r);
  pairs(r);
  const copy = {...r, '2':103};
  snapshot(copy);
  console.log(Object.keys(r).length, Object.keys(copy).length);
  pairs(copy);
} else if (mode === 'prototypes') {
  const r = {};
  const names = ['constructor','__defineGetter__','__defineSetter__','hasOwnProperty','__lookupGetter__','__lookupSetter__','isPrototypeOf','propertyIsEnumerable','toString','valueOf','__proto__','toLocaleString','missing','','toStringX'];
  for (const k of names) {
    console.log(+Object.hasOwn(r,k), +Object.hasOwn(r,k), +Object.hasOwn(r,k));
    if(k==='__proto__') define(r,k,7); else r[k]=7;
    console.log(+Object.hasOwn(r,k), +(k in r), r[k]);
    delete r[k];
    console.log(+Object.hasOwn(r,k), +Object.hasOwn(r,k));
  }
  snapshot(r);
} else if (mode === 'reads') {
  const r={hit:42}; define(r,'toString',7);
  for(const k of ['hit','toString']) console.log(r[k], +Object.hasOwn(r,k), +(k in r));
  for(const k of ['missing','valueOX','toStrinX','__proto_X','constructoX',
    'isPrototypeOX','hasOwnPropertX','toLocaleStrinX','__defineGetter_X',
    '__defineSetter_X','__lookupGetter_X','__lookupSetter_X','propertyIsEnumerablX',
    '', 'a-long-key-more-than-twenty-bytes','世界🌍']) {
    console.log(+(r[k]===undefined), +(k in r));
  }
  const reference={}; define(reference,'toString',undefined);
  console.log(+Object.hasOwn(reference,'toString'), +(reference.toString===undefined), +('toString' in reference));
  reference.toString='owned value'; const held=reference.toString;
  delete reference.toString;
  console.log(+Object.hasOwn(reference,'toString'), +Object.hasOwn(reference,'toString'));
  console.log(JSON.stringify(held));
} else if (mode === 'references') {
  const r = {};
  for (let i=0;i<10000;i++) r['same'] = 'value';
  const held = r.same, keys = Object.keys(r);
  delete r.same;
  console.log(JSON.stringify(held)); console.log(JSON.stringify(keys));
  const churn = {};
  for(let round=0;round<10;round++) {
    for(let i=0;i<1000;i++) churn[`k${i}`] = `k${i}`;
    const snapshot = Object.keys(churn);
    for(let i=0;i<1000;i++) delete churn[`k${i}`];
  }
  console.log(Object.keys(churn).length);
  const missingValue = {undefined:undefined};
  console.log(+Object.hasOwn(missingValue,'undefined'), +Object.hasOwn(missingValue,'undefined'), +(missingValue.undefined===undefined));
  const scalar = {flag:true};
  console.log(+scalar.flag); scalar.flag=false;
  console.log(+Object.hasOwn(scalar,'flag'), +scalar.flag);
  let chain=null;
  for(let i=0;i<100000;i++) chain={next:chain};
  chain=null;
  console.log('chain released');
} else if (mode === 'iteration') {
  const r = {'10':10,'2':2,a:1,b:2};
  // Ordinary for...in snapshots own keys on this unmodified-prototype object in V8.
  // Trigger mutation after the first yield, which is still '2', in both logical iterators.
  for (const k in r) {
    if(k==='2') { delete r['10']; r.new=3; r.a=9; }
    console.log(JSON.stringify(k)); console.log(r[k]);
  }
  // A second iterator with the original snapshot, also reading current values.
  for (const k of ['2','10','a','b']) {
    if(Object.hasOwn(r,k)) { console.log(JSON.stringify(k)); console.log(r[k]); }
  }
} else if (mode === 'numeric') {
  const r={}, n=100000;
  for(let i=n;i>0;i--) r[String(i*3)]=i;
  for(let i=2;i<=n;i+=2) delete r[String(i*3)];
  for(let i=n;i>0;i--) r[String(i*3)]=i;
  pairs(r);
} else if (mode === 'two-indices') {
  const r={};
  define(r,'10',0); define(r,'2',1);
  snapshot(r);
} else if (mode === 'bench' || mode === 'workload') {
  const n=Number(process.argv[3]), r={};
  const now=()=>Number(process.hrtime.bigint())/1e9;
  let start=now();
  for(let i=0;i<n;i++) r[`k${i}`]=i;
  const build=now()-start;
  start=now(); let sum=0;
  for(let i=0;i<n;i++) sum+=r[`k${i}`];
  const hit=now()-start;
  start=now(); let misses=0;
  for(let i=0;i<n;i++) misses+=r[`m${i}`]===undefined;
  const miss=now()-start;
  start=now();
  for(let i=0;i<n;i+=2) delete r[`k${i}`];
  const deletion=now()-start;
  start=now(); let visited=0;
  for(const k in r) {
    visited++; sum+=r[k];
    if(mode==='workload') console.log(k);
  }
  const iteration=now()-start;
  console.log(sum,misses,visited,Object.keys(r).length);
  if(mode==='bench') console.log(build.toFixed(9),hit.toFixed(9),miss.toFixed(9),deletion.toFixed(9),iteration.toFixed(9));
} else throw new Error(`unknown mode ${mode}`);
