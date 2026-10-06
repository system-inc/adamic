declare const process: {
    readonly exit: (code: number) => never;
};

console.log('one finding');
process.exit(1);
