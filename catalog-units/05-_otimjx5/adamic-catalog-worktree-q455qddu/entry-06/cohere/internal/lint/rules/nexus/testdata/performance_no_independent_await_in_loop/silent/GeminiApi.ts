// Shape from ahra modules/google/gemini/GeminiApi.ts:255 (waitForFileToBecomeActive).
// Polling: a `for` whose bound is an attempt budget rather than a collection's length, a
// setTimeout sleep between polls, and `current` reassigned from the awaited response.
interface GeminiFileInterface {
    name: string;
    state: 'PROCESSING' | 'ACTIVE' | 'FAILED';
}
declare function setTimeout(callback: (value?: unknown) => void, milliseconds: number): unknown;
declare function readErrorMessage(response: Response): Promise<string>;
declare const generativeLanguageApiBase: string;
interface Response {
    ok: boolean;
    status: number;
    json(): Promise<unknown>;
}
declare function fetch(url: string, init: { headers: Record<string, string> }): Promise<Response>;

export class GeminiApi {
    constructor(private readonly apiKey: string) {}

    async waitForFileToBecomeActive(file: GeminiFileInterface): Promise<GeminiFileInterface> {
        let current = file;
        for(let attempt = 0; attempt < 60 && current.state === 'PROCESSING'; attempt++) {
            await new Promise(function(resolve) {
                setTimeout(resolve, 2000);
            });
            const response = await fetch(`${generativeLanguageApiBase}/${current.name}`, {
                headers: { 'x-goog-api-key': this.apiKey },
            });
            if(!response.ok) {
                throw new Error(`Gemini file status check failed (${response.status}): ${await readErrorMessage(response)}`);
            }
            current = (await response.json()) as GeminiFileInterface;
        }
        if(current.state === 'FAILED') {
            throw new Error('Gemini could not process the uploaded audio file.');
        }
        return current;
    }
}
