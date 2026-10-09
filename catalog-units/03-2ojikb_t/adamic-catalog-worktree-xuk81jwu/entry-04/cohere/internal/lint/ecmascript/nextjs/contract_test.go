package nextjs

import "testing"

// The five phi web routes that lost prerendering are the first rows, by their real paths, because
// they are the files this helper exists for.
func TestIsRouteContractExport(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		filePath string
		export   string
		want     bool
	}{
		{"a dynamic blog page", "/www-phi-health/app/(main-layout)/blog/[slugWithId]/page.tsx", "generateStaticParams", true},
		{"a catch-all inside route groups", "/www-phi-health/app/(main-layout)/(shop)/shop/(products)/axis/[[...variant]]/page.tsx", "generateStaticParams", true},
		{"the near miss is not the contract", "/www-phi-health/app/(main-layout)/blog/[slugWithId]/page.tsx", "generateStaticParameters", false},

		{"segment configuration on a layout", "/repository/app/layout.tsx", "revalidate", true},
		{"the abbreviation gate's trap", "/repository/app/api/report/route.ts", "maxDuration", true},
		{"page metadata", "/repository/src/app/page.jsx", "generateMetadata", true},
		{"an http method on a route", "/repository/app/api/route.ts", "POST", true},
		// A page has no http methods and a route has no metadata, which is the types plugin's two arms.
		{"an http method on a page is not read", "/repository/app/page.tsx", "POST", false},
		{"metadata on a route is not read", "/repository/app/api/route.ts", "metadata", false},
		{"image metadata", "/repository/app/opengraph-image.tsx", "alt", true},
		{"a numbered image metadata file", "/repository/app/blog/icon2.tsx", "size", true},
		{"sitemaps", "/repository/app/sitemap.ts", "generateSitemaps", true},
		{"robots takes segment configuration only", "/repository/app/robots.ts", "generateSitemaps", false},
		// Read from the installed build: 16.2 dropped this one, and still reads the two that replaced it.
		{"a name the installed build no longer reads", "/repository/app/page.tsx", "experimental_ppr", false},
		{"its replacement", "/repository/app/page.tsx", "unstable_instant", true},

		// The `app` test is a segment test, unlike IsInApplicationDirectory's substring test, so a
		// checkout whose name ends in the word does not make every file in it a route.
		{"a checkout named for the word", "/www-connected-app/source/page.tsx", "revalidate", false},
		{"no app directory at all", "/repository/pages/index.tsx", "revalidate", false},
		{"a template exports only default", "/repository/app/template.tsx", "revalidate", false},
		{"a page with no module extension", "/repository/app/page.mdx", "revalidate", false},
		{"a file named app", "/repository/app", "revalidate", false},
		{"an empty path", "", "revalidate", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRouteContractExport(testCase.filePath, testCase.export); got != testCase.want {
				t.Errorf("IsRouteContractExport(%q, %q) = %v, want %v", testCase.filePath, testCase.export, got, testCase.want)
			}
		})
	}
}
