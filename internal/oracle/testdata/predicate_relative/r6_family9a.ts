interface Program { getName(): string }
interface State { tag: string; program: Program | undefined }
interface WithProgram { tag: string; program: Program }
function isWithProgram(state: State): state is WithProgram {
	return state.program !== undefined;
}
const state: State = { tag: "s", program: { getName: () => "p" } };
console.log(isWithProgram(state) ? state.program.getName() : "none");
