import { registerHooks } from 'node:module';
import { pathToFileURL } from 'node:url';
if (!process.env.PUBLIC_ANY_API) throw Error('PUBLIC_ANY_API must name the built TypeScript 6.0.3 API');
registerHooks({ resolve(specifier, context, next) {
    if (specifier === 'public-any-ts') return { url: pathToFileURL(process.env.PUBLIC_ANY_API).href, shortCircuit: true };
    return next(specifier, context);
} });
