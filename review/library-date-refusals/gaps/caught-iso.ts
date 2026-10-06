try {
    new Date(NaN).toISOString();
} catch (error) {
    if (error instanceof Error) console.log(error.name);
}
