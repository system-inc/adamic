const empty: never[] = [];
function update(values: number[]): void {
    values[0] = 0;
    console.log(`${values.length}`);
}
update(empty);
