export class Entry {
    readonly source: string; readonly replacement: string; readonly owner: string; readonly id: string; readonly message: string;
    constructor(source: string, replacement: string, owner: string, id: string, message: string) { this.source = source; this.replacement = replacement; this.owner = owner; this.id = id; this.message = message; }
}
const entries = [
    new Entry("next/navigation", "@structure/source/router/Navigation", "/source/router/", "forbiddenNavigationImport", "Importing from 'next/navigation' is not allowed. Use '@structure/source/router/Navigation', which is the same navigation without binding the file to one framework's router."),
    new Entry("next/link", "@structure/source/components/navigation/Link", "/source/components/navigation/Link.tsx", "forbiddenLinkImport", "Importing from 'next/link' is not allowed. Use '@structure/source/components/navigation/Link' instead, so a link renders the same way whichever framework is underneath it."),
    new Entry("next/image", "@structure/source/components/images/Image", "/source/components/images/Image.tsx", "forbiddenImageImport", "Importing from 'next/image' is not allowed. Use '@structure/source/components/images/Image' instead, so an image renders the same way whichever framework is underneath it."),
    new Entry("framer-motion", "motion/react", "", "forbiddenMotionImport", "Importing from 'framer-motion' is not allowed. Use 'motion/react', which is the same library under the name it ships as now."),
];
export function entry(source: string): Entry | undefined {
    for(const candidate of entries) { if(candidate.source === source) { return candidate; } }
    return undefined;
}
