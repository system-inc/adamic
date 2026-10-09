// A small structural range type also admits branches carrying children.
interface RangeInterface {
    readonly start: number;
    readonly end: number;
}
interface BranchInterface extends RangeInterface {
    readonly children: readonly RangeInterface[];
}
const branches: BranchInterface[] = [];
function append(children: readonly RangeInterface[]): void {
    branches.push({ start: 0, end: 1, children });
}
function ranges(): RangeInterface[] {
    const result: RangeInterface[] = [];
    result.push({ start: 0, end: 1 });
    return result;
}
append(ranges());
console.log(`${branches.length}`);
