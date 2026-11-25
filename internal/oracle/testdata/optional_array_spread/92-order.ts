interface State { values?: string[] }
function replace(state: State): string[] { state.values = ["new"]; return ["next"]; }
function add(state: State): void {
    const saved = state.values;
    state.values?.push(...replace(state));
    console.log(saved?.join(",") ?? "absent");
    console.log(state.values?.join(",") ?? "absent");
}
add({ values: ["first"] });
