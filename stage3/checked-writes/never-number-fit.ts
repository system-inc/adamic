const empty: never[] = [];
function update(values: number[]): void {
    console.log(`${values.length}`);
    values.splice(0, 0);
    values.fill(0);
    console.log(`${values.length}`);
}
update(empty);
