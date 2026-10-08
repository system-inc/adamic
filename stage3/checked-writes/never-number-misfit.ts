const empty: never[] = [];
function update(values: number[]): void {
    values.push(0);
    console.log(`${values.length}`);
}
update(empty);
