// Shape from ahra modules/samsung/frame-tv/FrameTvApi.ts:230 (isArtModeEnabled).
// Retry: an attempt counter bounded by maximumAttempts rather than a collection's length, a return
// on the first awaited success, and a log line per failed attempt.
declare const console: { log(...values: unknown[]): void };

export class FrameTvApi {
    async disconnect(): Promise<void> {}
    async connect(): Promise<void> {}
    async sendRequest(payload: object, event?: string, timeout?: number): Promise<unknown> {
        return { payload, event, timeout };
    }

    async isArtModeEnabled(): Promise<boolean> {
        const maximumAttempts = 3;
        const timeoutInMilliseconds = 5000;

        for(let attempt = 1; attempt <= maximumAttempts; attempt++) {
            try {
                await this.disconnect();
                await this.connect();
                const response = (await this.sendRequest(
                    { request: 'get_artmode_status' },
                    'art_app_request',
                    timeoutInMilliseconds,
                )) as { value: string };
                return response.value === 'on';
            }
            catch(error) {
                if(attempt === maximumAttempts) throw error;
                console.log(`[FrameApi] isArtModeEnabled attempt ${attempt} failed, retrying...`);
            }
        }
        throw new Error('Art mode check failed after all attempts');
    }
}
