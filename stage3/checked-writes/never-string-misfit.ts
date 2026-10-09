const empty: never[] = [];
function update(values: string[]): void {
    values.push("bad" + "bad");
    console.log(`${values.length}`);
}
update(empty);
