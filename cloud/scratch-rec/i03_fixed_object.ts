const fixed: { k: number } = { k: 41 };
function read(o: object): string {
  if ('k' in o) {
    return typeof o.k;
  }
  return 'absent';
}
console.log(read(fixed));
