const empty: never[] = [];
function update(values: { name: string }[]): void {
    values.push({ name: "new" + "new" });
    console.log(`${values.length}`);
}
update(empty);
