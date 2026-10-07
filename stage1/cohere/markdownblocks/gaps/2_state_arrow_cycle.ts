// The literal Go/JS state-machine closure shape; the state cell and its closure hold each other.
function run(): void {
    let remaining = 2;
    let state = (code: number): boolean => false;
    state = (code: number): boolean => {
        remaining--;
        console.log(`${code}`);
        return remaining > 0 && state(code);
    };
    state(35);
}
run();
