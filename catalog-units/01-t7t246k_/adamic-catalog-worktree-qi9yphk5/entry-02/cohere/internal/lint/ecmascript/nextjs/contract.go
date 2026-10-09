package nextjs

import (
	"strings"
)

// The names below are what Next.js reads off an App Router module BY NAME, with no import anywhere
// to say so. Read from the installed build, next 16.2.11, rather than from documentation:
//
//	routeSegmentConfiguration   the `checkFields` block next-types-plugin generates for every page,
//	                            layout and route (build/webpack/plugins/next-types-plugin/index.js:46)
//	pageMetadataExports         the same block's page-and-layout arm
//	HTTP methods                server/web/http.js `HTTP_METHODS`, the route arm of the same block
//	image metadata exports      build/webpack/loaders/next-metadata-image-loader.js and
//	                            next-metadata-route-loader.js
//	generateSitemaps            next-metadata-route-loader.js
//
// `experimental_ppr` is absent on purpose. Older documentation lists it, and 16.2 no longer reads it:
// the types plugin carries `unstable_instant` and `unstable_dynamicStaleTime` in its place. A list
// copied from documentation rather than read from the build would exempt a dead name and miss two
// live ones, which is why the build is the source.
//
// A name here is a contract rather than a choice. Renaming it, recasing it, or dropping its `export`
// produces a file that compiles, renders, and is never read for that name again, so the route quietly
// loses its static params, its metadata, or its revalidation. No error says so anywhere. Five phi web
// routes exported `generateStaticParameters` for fifteen months for exactly this reason.
var routeSegmentConfiguration = []string{
	"config",
	"generateStaticParams",
	"unstable_instant",
	"unstable_dynamicStaleTime",
	"revalidate",
	"dynamic",
	"dynamicParams",
	"fetchCache",
	"preferredRegion",
	"runtime",
	"maxDuration",
}

var pageMetadataExports = []string{"metadata", "generateMetadata", "viewport", "generateViewport"}

var httpMethodExports = []string{"GET", "HEAD", "OPTIONS", "POST", "PUT", "DELETE", "PATCH"}

var imageMetadataExports = []string{"alt", "size", "contentType", "generateImageMetadata"}

// routeFileContracts maps an App Router file's stem to the names Next reads from it.
//
// The UI convention files that export only `default` (template, loading, error, not-found and their
// siblings) have no named contract, so they are absent rather than mapped to nothing: `default` is not
// an identifier a naming rule can see.
var routeFileContracts = buildRouteFileContracts()

func buildRouteFileContracts() map[string]map[string]bool {
	set := func(groups ...[]string) map[string]bool {
		names := map[string]bool{}
		for _, group := range groups {
			for _, name := range group {
				names[name] = true
			}
		}
		return names
	}
	pageOrLayout := set(routeSegmentConfiguration, pageMetadataExports)
	imageMetadata := set(routeSegmentConfiguration, imageMetadataExports)
	return map[string]map[string]bool{
		"page":            pageOrLayout,
		"layout":          pageOrLayout,
		"route":           set(routeSegmentConfiguration, httpMethodExports),
		"icon":            imageMetadata,
		"apple-icon":      imageMetadata,
		"opengraph-image": imageMetadata,
		"twitter-image":   imageMetadata,
		"sitemap":         set(routeSegmentConfiguration, []string{"generateSitemaps"}),
		"robots":          set(routeSegmentConfiguration),
		"manifest":        set(routeSegmentConfiguration),
	}
}

// routeModuleExtensions are Next's default `pageExtensions`.
var routeModuleExtensions = []string{".tsx", ".ts", ".jsx", ".js"}

// RouteContractExports returns the export names Next.js reads by name from an App Router file, or
// nil when the path is not one.
//
// The file must sit under a path SEGMENT named `app`, which is deliberately tighter than
// `IsInApplicationDirectory`: that one is a substring test and answers true for every file in
// `www-connected-app/`, and here a false yes would silence naming rules across a whole checkout. A
// numbered metadata image (`icon1.tsx`, `opengraph-image2.tsx`) is the same convention as the bare
// name, so trailing digits are stripped from those stems before the lookup.
//
// The answer is a set the caller must not modify.
func RouteContractExports(filePath string) map[string]bool {
	segments := splitSegments(filePath)
	if len(segments) < 2 {
		return nil
	}
	underApplication := false
	for _, segment := range segments[:len(segments)-1] {
		if segment == "app" {
			underApplication = true
			break
		}
	}
	if !underApplication {
		return nil
	}

	stem := ""
	baseName := segments[len(segments)-1]
	for _, extension := range routeModuleExtensions {
		if strings.HasSuffix(baseName, extension) {
			stem = strings.TrimSuffix(baseName, extension)
			break
		}
	}
	if stem == "" {
		return nil
	}
	if contract, found := routeFileContracts[stem]; found {
		return contract
	}
	numberless := strings.TrimRight(stem, "0123456789")
	switch numberless {
	case "icon", "apple-icon", "opengraph-image", "twitter-image":
		return routeFileContracts[numberless]
	}
	return nil
}

// IsRouteContractExport reports whether Next.js reads this name from this file.
//
// Every rule that renames, recases or un-exports an identifier asks this before it reports or fixes,
// so the list lives once, here, rather than once per rule's options where it drifts. The rules'
// consumer options still exist for what only a consumer knows; this is the floor they cannot drop
// below.
func IsRouteContractExport(filePath string, name string) bool {
	return RouteContractExports(filePath)[name]
}
