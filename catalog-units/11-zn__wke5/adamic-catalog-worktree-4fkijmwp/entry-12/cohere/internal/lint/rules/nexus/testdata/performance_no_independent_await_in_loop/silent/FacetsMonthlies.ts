// Shape from ahra modules/facets/FacetsMonthlies.ts:302. The loop body only awaits, but the callee
// prints a multi-line progress block per month, so the order the loop imposes is the order the
// reader sees. Silent through the one-level callee scan; the console.error inside the catch is
// deliberately not what silences it (a catch is a failure path).
declare const console: { log(...values: unknown[]): void; error(...values: unknown[]): void };
declare function callModel(prompt: string): Promise<string>;
declare function getMonthsNeedingMonthlies(facet: string): Array<{ yearMonth: string; reason: string }>;

async function generateFacetMonthly(facet: string, yearMonth: string, options: { model: string }): Promise<void> {
    console.log(`\nGenerating monthly for ${yearMonth}...`);
    try {
        const output = await callModel(`${facet} ${yearMonth} ${options.model}`);
        console.log(`  Written ${output.length} characters`);
    }
    catch(error) {
        console.error(`  Monthly generation failed for ${yearMonth}`, error);
    }
}

export async function updateFacetMonthlies(facet: string, options: { model: string }): Promise<void> {
    const monthsNeeding = getMonthsNeedingMonthlies(facet);
    for(const month of monthsNeeding) {
        await generateFacetMonthly(facet, month.yearMonth, options);
    }
}
