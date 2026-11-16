function show(value: PromiseLike<number>): string {
  return typeof value;
}
console.log(show(Promise.resolve(7)));
