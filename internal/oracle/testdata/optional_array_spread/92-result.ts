interface State { values?: string[] }
function add(state: State, more: string[]): void {
    const n = state.values?.push(...more);
    console.log(String(n));
}
add({ values: ["first"] }, ["next"]);
add({}, ["next"]);
