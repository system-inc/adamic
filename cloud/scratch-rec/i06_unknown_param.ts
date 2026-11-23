const nums: Record<string, number> = { k: 41 };
function read(o: unknown): string {
  if (typeof o === 'object' && o !== null && 'k' in o) {
    return typeof o.k;
  }
  return 'absent';
}
console.log(read(nums));
