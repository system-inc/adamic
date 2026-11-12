enum Extension { Ts = '.ts', Js = '.js' }
function probe(flatBuiltins: readonly Extension[], extension: string): boolean {
    return !(flatBuiltins as readonly string[]).includes(extension);
}
console.log(`${probe([Extension.Ts], '.ts')}`);
