const empty: never[] = [];
const view: number[] = empty;
const ordinary: number[] = [];
function update(values: number[]): void {
    values.push(0);
    console.log(`${values.length}`);
}
console.log(`${view.length}`);
update(ordinary);
