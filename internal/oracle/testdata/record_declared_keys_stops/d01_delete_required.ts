interface Options { target?: number; required: number; [option: string]: number | undefined; }
function dropRequired(options: Options, key: string): void { delete options[key]; }
const options: Options = { required: 4 };
dropRequired(options, "required");
console.log(typeof options.required);
