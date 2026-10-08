const empty: never[] = [];
function update(values: { name: string }[]): void {
    console.log(`${values.length}`);
    values.splice(0, 0);
    values.fill({ name: "new" + "new" });
    console.log(`${values.length}`);
}
update(empty);
