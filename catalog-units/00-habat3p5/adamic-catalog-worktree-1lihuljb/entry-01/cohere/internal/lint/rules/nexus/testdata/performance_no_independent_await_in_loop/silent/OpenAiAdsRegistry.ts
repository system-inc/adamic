// Shape from ahra modules/openai/ads/OpenAiAdsRegistry.ts:379 (openAiAdsPaginate).
// Cursor pagination in a `while (true)`, and the inner for-of over a page holds no await, only a
// yield. Neither loop is a candidate.
interface OpenAiAdsListResponseInterface<ItemType> {
    data?: ItemType[];
    hasMore?: boolean;
    lastId?: string;
}
interface OpenAiAdsPaginateOptionsInterface {
    after?: string;
    limit?: number;
    order?: string;
    filters?: Record<string, string>;
    apiKey?: string;
}
declare function openAiAdsRequest<ResponseType>(
    path: string,
    options: { method: string; query: Record<string, unknown>; apiKey?: string },
): Promise<ResponseType>;

export async function* openAiAdsPaginate<ItemType>(
    path: string,
    options: OpenAiAdsPaginateOptionsInterface = {},
): AsyncGenerator<ItemType, void, void> {
    let cursor: string | undefined = options.after;
    const pageLimit = options.limit;

    while(true) {
        const query: Record<string, string | number | boolean | undefined | null> = {
            ...(options.filters ?? {}),
        };
        if(pageLimit !== undefined) query['limit'] = pageLimit;
        if(options.order !== undefined) query['order'] = options.order;
        if(cursor !== undefined) query['after'] = cursor;

        const response = await openAiAdsRequest<OpenAiAdsListResponseInterface<ItemType>>(path, {
            method: 'GET',
            query,
            apiKey: options.apiKey,
        });

        const data = response.data ?? [];
        for(const item of data) {
            yield item;
        }

        if(response.hasMore === false) return;
        if(response.hasMore === undefined) {
            if(pageLimit === undefined) return;
            if(data.length < pageLimit) return;
        }

        if(!response.lastId) return;
        cursor = response.lastId;
    }
}
