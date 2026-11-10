import 'node:process';
Error.captureStackTrace = (target: object): void => { console.log(String(Version.later.length)); };
class Version {
	static zero = new Version();
	static later: string = "set";
	label: string;
	constructor() {
		const target = new Error("t");
		Error.captureStackTrace(target);
		this.label = "v";
	}
}
console.log(Version.zero.label);
