package nextjs

import "testing"

func TestAdamicAllHelpers(t *testing.T) {
	for _, path := range []string{"", "pages", "pages/_documentation.tsx", "pages\\_document.tsx", "/世界/😀/pages/_document/index.tsx", "pages/api/pages/x.ts", "src/myapp/x", "app/icon2.tsx", "app/iconx.ts", "app/page.mdx", "app/page.tsx", "app/route.ts", "app/sitemap.js", "app/robots.js", "app/manifest.ts", "app/apple-icon7.jsx", "app/opengraph-image9.ts", "app/twitter-image0.ts", "//app///layout.ts", "myapp/page.tsx", "app", "pages/", "pages/apiary/x.ts"} {
		IsDocumentFile(path)
		lastSeparator(path)
		splitPath(path)
		IsDocumentPage(path)
		IsInApplicationDirectory(path)
		IsInPagesDirectory(path)
		splitSegments(path)
		RouteContractExports(path)
	}
	buildRouteFileContracts()
}
