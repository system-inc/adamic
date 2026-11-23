const nums: Record<string, number> = { k: 41 };
function read(o: object): string {
  if ('k' in o) { return String(o.k); }
  return 'absent';
}
console.log(read(nums));
