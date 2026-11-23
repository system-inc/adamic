// A record whose values are numbers or booleans keeps them unboxed in its slots, so a dynamic read of
// one, past an `in` test on a receiver seen as object or unknown, boxes the value by the record's kind
// (union.c's adamic_record_value_union), never retains it as a pointer (#3tjjcjs, Language's i01 to
// i06 on #859ksab). The string record, the fixed object and the absent key are the controls.
const nums: Record<string, number> = { k: 41 };
const strs: Record<string, string> = { k: 'x' };
const fixed: { k: number } = { k: 41 };
const absent: Record<string, number> = { j: 1 };
const flags: Record<string, boolean> = { k: true };

function kind(o: object): string {
  if ('k' in o) {
    return typeof o.k;
  }
  return 'absent';
}

function text(o: object): string {
  if ('k' in o) {
    return String(o.k);
  }
  return 'absent';
}

function unknownKind(o: unknown): string {
  if (typeof o === 'object' && o !== null && 'k' in o) {
    return typeof o.k;
  }
  return 'absent';
}

console.log(kind(nums) + ' ' + kind(strs) + ' ' + kind(fixed) + ' ' + kind(absent) + ' ' + kind(flags));
console.log(text(nums) + ' ' + text(strs) + ' ' + text(fixed) + ' ' + text(absent) + ' ' + text(flags));
console.log(unknownKind(nums) + ' ' + unknownKind(flags) + ' ' + unknownKind(absent));
