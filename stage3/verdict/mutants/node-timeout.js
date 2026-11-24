// Verdict case: 001_varianceCantBeStrictWhileStructureIsnt, Node timeout.
if (process.env.STAGE3_VERDICT_SUITE === 'acceptance'
    && process.env.STAGE3_VERDICT_CASE === '001_varianceCantBeStrictWhileStructureIsnt'
    && process.env.STAGE3_VERDICT_SIDE === 'reference') {
    setTimeout(() => {}, 195_000);
} else {
    require(process.env.STAGE3_VERDICT_NODE_TSC);
}
