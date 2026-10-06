declare const process: {
    readonly env: { readonly NO_COLOR: string | undefined };
};

console.log(process.env.NO_COLOR === undefined ? 'color permitted' : 'color disabled');
