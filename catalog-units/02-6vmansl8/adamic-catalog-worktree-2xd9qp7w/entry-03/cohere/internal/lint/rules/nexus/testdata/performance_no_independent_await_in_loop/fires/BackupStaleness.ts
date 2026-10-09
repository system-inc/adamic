// Shape from ahra modules/backup/BackupStaleness.ts:367 (checkAllTrees) and :378 (checkAllCorpora).
// Two loops, each checking one independent tree or corpus and pushing the row. The sort happens
// after the loop, so the push order is irrelevant and Promise.all returns the same rows.
// Expected: 2 findings, one per loop.
interface TreeCoverageInterface {
    tree: string;
    status: 'Empty' | 'Partial' | 'Covered';
}
interface CorpusFreshnessInterface {
    corpus: string;
    status: 'Empty' | 'Stale' | 'Unchanged' | 'Fresh';
}
declare const monitoredTrees: string[];
declare const monitoredCorpora: string[];
declare function checkTreeCoverage(tree: string): Promise<TreeCoverageInterface>;
declare function checkCorpusFreshness(target: string): Promise<CorpusFreshnessInterface>;

// Check every monitored tree, worst first.
export async function checkAllTrees(): Promise<TreeCoverageInterface[]> {
    const results: TreeCoverageInterface[] = [];
    for(const tree of monitoredTrees) {
        results.push(await checkTreeCoverage(tree));
    }
    const rank = { Empty: 0, Partial: 1, Covered: 2 };
    return results.sort((first, second) => rank[first.status] - rank[second.status]);
}

// Check every monitored corpus. Returns one row each, worst first, so a reader
// sees the problem before the reassurance.
export async function checkAllCorpora(): Promise<CorpusFreshnessInterface[]> {
    const results: CorpusFreshnessInterface[] = [];
    for(const target of monitoredCorpora) {
        results.push(await checkCorpusFreshness(target));
    }
    const rank = { Empty: 0, Stale: 1, Unchanged: 2, Fresh: 3 };
    return results.sort((first, second) => rank[first.status] - rank[second.status]);
}
