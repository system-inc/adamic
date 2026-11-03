// A dictionary read the ordinary way, not a weak one: a computed key put 3 under mode, typed 1 | 2,
// and Node prints it. typeof says number, so before the literal-membership test this read 3 as a
// 1 | 2 natively; now it stops, naming the read and the 3.
interface Options { mode?: 1 | 2; [name: string]: string | number | undefined }
const name: string = "mode";
const options: Options = { [name]: 3 };
console.log("read");
console.log(String(options.mode));
