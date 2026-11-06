// The sameKeeping condition on #6c4swvt: a function over Options, whose values are boxed as a union, seen as one over a
// record of strings, held as strings. The view drops strict, but the index value representations differ, so it stops
// not yet at the existing storage stop, and the function never reads a string slot as a union box.
interface Options {
    strict?: boolean;
    [option: string]: string | boolean | undefined;
}
const show: (options: { [option: string]: string }) => string = (options: Options) => String(options.strict);
console.log(show({ outDir: "out" }));
