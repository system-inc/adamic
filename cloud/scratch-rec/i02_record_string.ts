const strs: Record<string, string> = { k: 'x' };
function read(o: object): string {
  if ('k' in o) {
    return typeof o.k;
  }
  return 'absent';
}
console.log(read(strs));
