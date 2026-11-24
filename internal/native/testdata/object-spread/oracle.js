// Node's answer for each of harness.c's modes, the same witness written as the program it stands for:
// each result's keys in Object.keys order, each key's value by Reflect.get, an in test where the program
// has one, and the source after the spread.
'use strict';
const mode = process.argv[2];

function show(value) {
  if (value === undefined) return 'undefined';
  if (value === null) return 'null';
  if (typeof value === 'function') return 'function:' + show(value());
  if (Array.isArray(value)) return 'array:' + value.join(',');
  if (typeof value === 'object') return 'object';
  return typeof value + ':' + String(value);
}

function print(label, object) {
  const keys = Object.keys(object);
  console.log(keys.length === 0 ? label : `${label} ${keys.join(',')}`);
  for (const key of keys) console.log(` ${key}=${show(Reflect.get(Object(object), key, object))}`);
}

// The modes whose program prints JSON.stringify of its result.
const jsonModes = new Set(['optional_string_absent', 'p01', 'p01_reverse', 'resolver_2774_absent', 'absent_kept']);

function witness(source, spread, probe) {
  const result = spread(source);
  print('result', result);
  if (probe !== undefined) console.log(`has ${probe} ${probe in result}`);
  if (jsonModes.has(mode)) console.log(`json ${JSON.stringify(result)}`);
  if (source !== undefined && source !== null) {
    print('source', source);
    console.log('source unchanged');
  }
  return result;
}

const undefinedK = (source) => ({ ...source, k: undefined });
const resolved = (source) => ({ ...source, path: '/b.ts', originalPath: '/c.ts' });
const resolved140 = () => ({ path: '/a.ts', extension: '.ts', packageId: undefined, resolvedUsingTsExtension: false });
const resolved2774 = () => ({ path: '/a.ts', extension: '.ts', packageId: undefined, originalPath: undefined, resolvedUsingTsExtension: false });

function region() {
  const source = { held: { held: undefined, label: 'n' }, label: 's' };
  const result = witness(source, (s) => ({ ...s, count: 1 }));
  console.log(`held.label=${show(result.held.label)}`);
  console.log(`held.label=${show(result.held.label)}`);
}

const modes = {
  number_source: () => witness({ k: 1 }, undefinedK, 'k'),
  number_source_undefined: () => witness(undefined, undefinedK, 'k'),
  number_source_null: () => witness(null, undefinedK, 'k'),
  number_present: () => witness({ k: 1 }, undefinedK, 'k'),
  optional_string_present: () => witness({ k: 'one' }, undefinedK, 'k'),
  optional_string_absent: () => witness({}, undefinedK, 'k'),
  optional_number_present: () => witness({ k: 1 }, undefinedK, 'k'),
  optional_number_absent: () => witness({}, undefinedK, 'k'),
  string_source: () => witness({ k: 'one' }, undefinedK, 'k'),
  string_source_undefined: () => witness(undefined, undefinedK, 'k'),
  width: () => witness({ a: 1, k: 'x' }, (source) => ({ ...source, k: 5 })),
  width_method: () => witness({ a: 1, k: 'x' }, (source) => ({ ...source, k() { return 5; } })),
  p01: () => witness({ kind: 'Holder', value: 1234.5678 }, (source) => ({ ...source, value: undefined }), 'value'),
  p01_true: () => witness({ kind: 'Holder', value: 1234.5678 }, (source) => ({ ...source, value: true }), 'value'),
  p01_reverse: () => witness({ kind: 'Holder', value: 'v' }, (source) => ({ ...source, value: 1234.5678 }), 'value'),
  absent_kept: () => witness({ a: 'x' }, (source) => ({ ...source, c: 1 }), 'b'),
  resolver_140: () => witness(resolved140(), resolved),
  resolver_2774_undefined: () => witness(resolved2774(), resolved),
  resolver_2774_absent: () => witness(resolved140(), resolved),
  cache_279: () => witness(
    { resolvedModule: { resolvedFileName: '/m.ts' }, affectingLocations: ['/x'] },
    (source) => ({ ...source, failedLookupLocations: ['/f'], affectingLocations: ['/g'], resolutionDiagnostics: [] }),
  ),
  intern: () => {
    print('result', undefinedK({ k: 1 }));
    print('result', undefinedK({ k: 1 }));
  },
  loop: () => {
    const source = { k: 1 };
    let result;
    for (let index = 0; index < 1000; index++) result = undefinedK(source);
    print('result', result);
  },
  freed_source: () => print('result', { ...{ k: 1, label: 's' }, k: 2 }),
  graph_source: region,
  program_source: region,
  graph_result_complete: () => print('result', undefinedK({ k: 1 })),
  graph_result_written: () => print('result', resolved(resolved140())),
  graph_result_absent: () => print('result', { ...{ a: 'x' }, c: 1 }),
};

if (!(mode in modes)) throw new Error(`no mode ${mode}`);
modes[mode]();
