// Recursive State itself, without captured locals or nested function declarations.
type StateType = (code: number) => StateType | undefined;
function step(code: number): StateType | undefined {
    console.log(`${code}`);
    return code > 0 ? step : undefined;
}
const next: StateType = step;
const last = next(1);
if(last !== undefined) last(0);
