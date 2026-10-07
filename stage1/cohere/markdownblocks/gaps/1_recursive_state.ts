// micromark/types.go: type State func(code Code) State; nil means finished.
type StateType = (code: number) => StateType | undefined;
function createState(): StateType {
    let remaining = 2;
    function next(code: number): StateType | undefined {
        console.log(`${code}`);
        remaining--;
        return remaining > 0 ? next : undefined;
    }
    return next;
}
let state: StateType | undefined = createState();
while(state !== undefined) state = state(35);
