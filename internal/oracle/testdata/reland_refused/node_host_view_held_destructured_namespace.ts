import process from "node:process";
const { config } = process;
console.log(String(config.target_defaults.cflags.length));
