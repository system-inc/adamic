const nums: Record<string, number> = { j: 1 };
function read(o: object): string {
  if ('k' in o) {
    return typeof o.k;
  }
  return 'absent';
}
console.log(read(nums));
