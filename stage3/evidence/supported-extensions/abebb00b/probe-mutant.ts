enum Extension { Ts = '.ts', Js = '.js' }
function probe(flatBuiltins: readonly Extension[], extension: string): boolean {
    return !flatBuiltins.includes(extension as Extension);
}
console.log(`${probe([Extension.Ts], '.ts')}`);
