interface Root { readonly kind: 'Holder'; }
interface View extends Root { readonly marker: string | number; }
const raw = { kind: 'Holder' as const, marker: 'ok', nullSlot: null, undefinedSlot: undefined, uint8Slot: new Uint8Array(1), int32Slot: new Int32Array(1), float64Slot: new Float64Array(1) };
const base: Root = raw;
const marker = (base as View).marker;
console.log(typeof marker === 'string' ? marker : 'wrong');
raw.uint8Slot[0] = 41;
raw.int32Slot[0] = -42;
raw.float64Slot[0] = 3.5;
console.log(`${raw.nullSlot === null}:${raw.undefinedSlot === undefined}`);
console.log(`${raw.uint8Slot[0]}:${raw.int32Slot[0]}:${raw.float64Slot[0]}`);
