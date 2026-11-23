const nums: Record<string, number> = { k: 41 };
function read<T extends object>(o: T): string {
  if ('k' in o) { return typeof o.k; }
  return 'absent';
}
console.log(read(nums));
