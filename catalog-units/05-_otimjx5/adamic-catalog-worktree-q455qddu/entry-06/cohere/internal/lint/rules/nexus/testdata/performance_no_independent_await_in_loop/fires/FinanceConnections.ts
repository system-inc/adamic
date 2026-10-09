// Shape from ahra modules/finance/connections/FinanceConnections.ts:70 (importAll).
// One adapter per source. The unauthorized branch records and continues, the catch records and moves
// on, nothing is printed, and every write lands in `results`, which the body never reads back.
// importTransactions and importBalances are synchronous idempotent importers keyed by source, so
// running the adapters together interleaves only whole synchronous imports.
// Expected: 1 finding.
interface TransactionSummaryInterface {
    unmappedAccounts: string[];
}
interface SourceImportResultInterface {
    label: string;
    ranTransactions: boolean;
    transactionSummary: TransactionSummaryInterface | null;
    balancesInserted: number;
    unmappedAccounts: string[];
    skippedReason: string | null;
}
interface FinanceAdapterInterface {
    label: string;
    source: string;
    isAuthorized(): boolean;
    fetchTransactions(sinceDate: string): Promise<unknown[]>;
    fetchBalances(): Promise<unknown[]>;
}
declare class AdapterNotAuthorizedError extends Error {}
declare function importTransactions(transactions: unknown[]): TransactionSummaryInterface;
declare function importBalances(balances: unknown[]): { inserted: number; unmappedAccounts: string[] };

export class FinanceConnections {
    constructor(private readonly adapters: FinanceAdapterInterface[]) {}

    async importAll(sinceDate: string): Promise<SourceImportResultInterface[]> {
        const results: SourceImportResultInterface[] = [];
        for(const adapter of this.adapters) {
            if(!adapter.isAuthorized()) {
                results.push({
                    label: adapter.label,
                    ranTransactions: false,
                    transactionSummary: null,
                    balancesInserted: 0,
                    unmappedAccounts: [],
                    skippedReason: 'not authorized -- waiting on Kirk to link this source',
                });
                continue;
            }
            try {
                const transactions = await adapter.fetchTransactions(sinceDate);
                const transactionSummary = importTransactions(transactions);
                const balances = await adapter.fetchBalances();
                const balanceResult = importBalances(balances);
                results.push({
                    label: adapter.label,
                    ranTransactions: true,
                    transactionSummary,
                    balancesInserted: balanceResult.inserted,
                    unmappedAccounts: [
                        ...new Set([...transactionSummary.unmappedAccounts, ...balanceResult.unmappedAccounts]),
                    ],
                    skippedReason: null,
                });
            }
            catch(error) {
                const skippedReason =
                    error instanceof AdapterNotAuthorizedError
                        ? error.message
                        : `error during import: ${(error as Error).message}`;
                results.push({
                    label: adapter.label,
                    ranTransactions: false,
                    transactionSummary: null,
                    balancesInserted: 0,
                    unmappedAccounts: [],
                    skippedReason,
                });
            }
        }
        return results;
    }
}
