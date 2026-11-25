interface State { values?: string[] }
function add(state: State): void {
    const saved = state.values;
    state.values?.push(...["next"]);
    console.log(String(saved === state.values));
    console.log(String(saved?.length));
}
add({ values: ["first"] });
