// Shape from ahra modules/os/reports/AhraOsReportPdf.ts:455. One shared PDF document gains a page
// per slide (`document.addPage`, a sink by its collection-building prefix), and the same document
// is handed to the await. The body reads back what it fills, so the pages are built in order.
interface PageInterface {
    drawRectangle(options: { x: number; y: number; width: number; height: number }): void;
}
interface PdfDocumentInterface {
    addPage(size: [number, number]): PageInterface;
}
interface SlideInterface {
    image: string | null;
}
declare function embedSlideHero(document: PdfDocumentInterface, image: string | null): Promise<{ width: number } | null>;
declare function drawSlide(page: PageInterface, slide: SlideInterface, hero: { width: number } | null): void;

export async function renderSlides(document: PdfDocumentInterface, slides: SlideInterface[]): Promise<void> {
    for(const slide of slides) {
        const page = document.addPage([1920, 1080]);
        page.drawRectangle({ x: 0, y: 0, width: 1920, height: 1080 });
        const hero = await embedSlideHero(document, slide.image);
        drawSlide(page, slide, hero);
    }
}
