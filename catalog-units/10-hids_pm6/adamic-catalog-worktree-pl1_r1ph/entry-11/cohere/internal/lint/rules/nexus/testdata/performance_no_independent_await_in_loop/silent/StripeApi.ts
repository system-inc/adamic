// Shape from ahra modules/finance/connections/stripe/StripeApi.ts:136 (list).
// Pagination: each page's starting_after comes from the page before it, and has_more decides
// whether there is a next page at all. A `while`, so never a candidate.
interface StripeResourceInterface {
    id: string;
}
interface StripeListResponseInterface {
    data: StripeResourceInterface[];
    has_more: boolean;
}
declare function request(apiKey: string, method: string, path: string): Promise<unknown>;

export async function list(apiKey: string, endpoint: string, parameters = {}): Promise<StripeResourceInterface[]> {
    const allItems: StripeResourceInterface[] = [];
    let hasMore = true;
    let startingAfter: string | null = null;

    while(hasMore) {
        const queryParameters: Record<string, string> = { limit: '100', ...parameters };
        if(startingAfter) queryParameters.starting_after = startingAfter;
        const query = new URLSearchParams(queryParameters).toString();
        const page = (await request(apiKey, 'GET', `${endpoint}?${query}`)) as unknown as StripeListResponseInterface;
        allItems.push(...page.data);
        hasMore = page.has_more;
        if(page.data.length > 0) {
            const lastItem = page.data[page.data.length - 1];
            if(lastItem) {
                startingAfter = lastItem.id;
            }
        }
    }
    return allItems;
}
