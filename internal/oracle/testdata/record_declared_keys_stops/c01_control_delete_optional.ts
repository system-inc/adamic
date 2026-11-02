interface Options { target?: number; required: number; [option: string]: number | undefined; }
function drop(options: Options, key: "target" | "extra"): void { delete options[key]; }
const options: Options = { target: 7, required: 4, extra: 9 };
drop(options, "extra");
console.log(`${options.required}:${options.target}:${"extra" in options}:${Object.keys(options).join(",")}`);
delete options.target;
console.log(`${options.required}:${typeof options.target}:${"target" in options}:${Object.keys(options).join(",")}`);
