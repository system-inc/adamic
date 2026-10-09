// commandLineParser.ts:2297 unknown recovery, extended with a stale nominal narrowing.
let value: unknown = new TypeError(`initial${1}`);
function change(): void { value = new Error(`replacement${2}`); }
try {
    if (value instanceof TypeError) {
        change();
        console.log(value.name);
    }
}
catch { console.log('caught'); }
finally { console.log('finally'); }
