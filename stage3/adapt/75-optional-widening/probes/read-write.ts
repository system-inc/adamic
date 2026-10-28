interface Narrow { x: number; }
interface Wide extends Narrow { p?: number; q?: number; }
interface Other { p?: string; }
const raw: Narrow = { x: 1 };
const writeView = raw as Wide;
const a = writeView;
let b: Wide = a;
function put(value: Wide): void { const c = value; c.p = 2; }
put(b);
const readView = raw as Wide;
console.log(readView.p);
const other: Other = {};
const otherAlias = other;
otherAlias.p = 'unrelated';
function apply(value: Wide, callback: (q: Wide) => void): void {
    const callbackAlias = callback;
    callbackAlias(value);
}
const callbackView = raw as Wide;
apply(callbackView, q => { const alias = q; alias.p = 3; });
const otherPropertyView = raw as Wide;
function putQ(value: Wide): void { const alias = value; alias.q = 4; }
putQ(otherPropertyView);
const computedView = raw as Wide;
function putKey(value: Wide, key: keyof Wide): void { const alias = value; alias[key] = 5; }
putKey(computedView, 'p');
const assignView = raw as Wide;
const assignAlias = assignView;
Object.assign(assignAlias, { p: 6 });
const untypedView = raw as Wide;
const untypedAlias = untypedView as any;
untypedAlias.p = 7;
const returnedView = raw as Wide;
const returnedAlias = Object.assign(returnedView, {});
returnedAlias.p = 8;
