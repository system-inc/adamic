const nums: Record<string, number> = { k: 41 };
function call(o: object): string {
  if ('k' in o && typeof o.k === 'function') {
    return String(o.k());
  }
  return 'not callable';
}
console.log(call(nums));
