package next

import (
	"crypto/sha256"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// Every expectation in this file was measured against the installed @next/eslint-plugin-next 16.3.1,
// run through ESLint 10.8.1's Linter with its working directory at a fixture tree laid out the same
// way, one fresh tree per case so the plugin's process-wide directory caches could not carry between
// them. Where upstream throws instead of answering, the row says so and pins what the port does.

// htmlLinkRow is one anchor and the hrefs it reports, once per matching route, as upstream names them.
type htmlLinkRow struct {
	element string
	reports []string
}

// htmlLinkRowsSource puts each row's element on a line of its own, from the third line on, so a
// finding's line says which row it belongs to.
func htmlLinkRowsSource(rows []htmlLinkRow) string {
	var source strings.Builder
	source.WriteString("export const C = () => (\n<>\n")
	for _, row := range rows {
		source.WriteString(row.element + "\n")
	}
	source.WriteString("</>\n);\n")
	return source.String()
}

// expectHtmlLinkRows checks every row's findings by line: how many, and that each carries the message
// for the href it names.
func expectHtmlLinkRows(t *testing.T, result rule_testing.Result, rows []htmlLinkRow) {
	t.Helper()
	text := result.SourceFile.Text()
	got := make(map[int][]string, len(rows))
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Message.Id != messageNoHtmlLinkForPages.Id {
			t.Errorf("a finding carries message id %q, want %q", diagnostic.Message.Id, messageNoHtmlLinkForPages.Id)
		}
		line := strings.Count(text[:diagnostic.Range.Pos()], "\n") - 2
		got[line] = append(got[line], diagnostic.Message.Description)
	}
	for index, row := range rows {
		want := make([]string, 0, len(row.reports))
		for _, hrefPath := range row.reports {
			want = append(want, noHtmlLinkForPagesMessage(hrefPath).Description)
		}
		if !slices.Equal(got[index], want) {
			t.Errorf("%s: reports %q, want %q", row.element, got[index], want)
		}
		delete(got, index)
	}
	for line, descriptions := range got {
		t.Errorf("findings on line %d, which holds no row: %q", line, descriptions)
	}
}

// htmlLinkTree is a fixture tree: every named file, plus the subject at the root.
func htmlLinkTree(rows []htmlLinkRow, paths ...string) map[string]string {
	files := map[string]string{"foo.tsx": htmlLinkRowsSource(rows)}
	for _, path := range paths {
		files[path] = "export {};"
	}
	return files
}

// runHtmlLinkTree runs the rule on a tree with no option set.
func runHtmlLinkTree(t *testing.T, rows []htmlLinkRow, paths ...string) {
	t.Helper()
	result := rule_testing.RunTypedFiles(t, NoHtmlLinkForPages, htmlLinkTree(rows, paths...), "foo.tsx")
	expectHtmlLinkRows(t, result, rows)
}

// runHtmlLinkTreeWithOption runs the rule on a tree on disk, with its option decoded as a config at the
// tree's root would decode it. option is the JSON written, with `{root}` standing for the tree's
// absolute directory.
func runHtmlLinkTreeWithOption(t *testing.T, rows []htmlLinkRow, option string, paths ...string) {
	t.Helper()
	result := rule_testing.RunTypedFilesWithOptionsFor(t, NoHtmlLinkForPages, htmlLinkTree(rows, paths...), "foo.tsx", func(directory string) any {
		quoted, _ := json.Marshal(directory)
		raw := strings.ReplaceAll(option, "{root}", strings.Trim(string(quoted), `"`))
		decoded, err := decodeNoHtmlLinkForPagesOptions([]byte(raw), rule.OptionsBase{ConfigDirectory: directory, ProjectRoot: directory})
		if err != nil {
			t.Fatalf("decoding %s: %v", raw, err)
		}
		return decoded
	})
	expectHtmlLinkRows(t, result, rows)
}

// upstreamRows are the anchors @next's own test file writes, with the hrefs its rows assert and three
// more measured on the same trees. Upstream asserts only the first finding of each; the counts here are
// every finding, measured.
func upstreamRows(home, dynamic, deepDynamic, profile, list, intercepted []string) []htmlLinkRow {
	return []htmlLinkRow{
		{`<Link href='/'>Homepage</Link>`, nil},
		{`<a>Homepage</a>`, nil},
		{`<a href='#heading'>Homepage</a>`, nil},
		{`<a href='https://google.com/'>Homepage</a>`, nil},
		{`<a href='/static-file.csv' download>Download</a>`, nil},
		{`<a target="_blank" href='/new-tab'>New Tab</a>`, nil},
		{`<a href='/presentation.pdf'>View PDF</a>`, nil},
		{`<a href='/'>Homepage</a>`, home},
		{`<a href='/list/foo/bar'>Homepage</a>`, prefix("/list/foo/bar/", deepDynamic)},
		{`<a href='/list/foo/'>Homepage</a>`, prefix("/list/foo/", dynamic)},
		{`<a href='/list/lorem-ipsum/'>Homepage</a>`, prefix("/list/lorem-ipsum/", dynamic)},
		{`<a href='/photo/1/'>Photo</a>`, intercepted},
		{`<a href='/profile'>Profile</a>`, profile},
		{`<a href='/list'>List</a>`, list},
	}
}

// prefix repeats one href once per entry of counts, which is how the rows spell a repeated finding.
func prefix(hrefPath string, counts []string) []string {
	repeated := make([]string, len(counts))
	for index := range counts {
		repeated[index] = hrefPath
	}
	return repeated
}

// The trees are upstream's fixture directories, test/unit/eslint-plugin-next at the 16.3.1 tag, file
// for file. Their contents are never read, only their names.
var (
	withCustomPagesDirectory = []string{"custom-pages/index.jsx", "custom-pages/[profile]/index.tsx", "custom-pages/list/[foo]/[id].jsx"}
	withAppDirectory         = []string{
		"app/page.jsx", "app/[profile]/page.tsx", "app/@modal/default.tsx", "app/@modal/(..)photo/page.tsx",
		"app/@modal/(..)photo/[id]/page.tsx", "app/list/[foo]/[id].tsx", "app/photo/[id]/page.tsx",
	}
	withoutPagesDirectory = []string{"index.jsx"}
)

func TestNoHtmlLinkForPagesReplaysUpstreamRows(t *testing.T) {
	t.Parallel()

	two, three := []string{"x", "x"}, []string{"x", "x", "x"}

	// 'valid link element', 'valid anchor element', ..., 'invalid static route' and 'invalid dynamic
	// route', with the custom pages directory written absolute, as upstream writes it. `/list/foo/`
	// matches both `[profile]/` and `list/[foo]/[id]`, so it reports twice.
	t.Run("custom pages directory", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTreeWithOption(t, upstreamRows([]string{"/"}, two, two, []string{"/profile/"}, []string{"/list/"}, []string{"/photo/1/"}),
			`"{root}/custom-pages"`, withCustomPagesDirectory...)
	})

	// 'valid link element with multiple directories'. The second directory sits inside the first and
	// is walked again from its own root, so `list/[foo]/[id]` is also the route `/[foo]/[id]`.
	t.Run("multiple custom pages directories", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTreeWithOption(t, upstreamRows([]string{"/"}, three, three, prefix("/profile/", two), prefix("/list/", two), prefix("/photo/1/", two)),
			`["{root}/custom-pages", "{root}/custom-pages/list"]`, withCustomPagesDirectory...)
	})

	// Every 'with appDir' row, the intercepted routes among them. `/` reports twice: `page.jsx` is
	// `^/$`, and `[profile]/page.tsx` is `/[profile]`, whose wildcard matches the empty segment. A
	// static app route never matches, because app routes have no trailing slash and hrefs gain one;
	// `/list` reports only through `[profile]`.
	t.Run("app directory", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, upstreamRows(prefix("/", two), two, two, []string{"/profile/"}, prefix("/list/", two), prefix("/photo/1/", two)),
			withAppDirectory...)
	})

	// 'prints warning when there are no "pages" or "app" directories': upstream warns once and
	// returns an empty visitor, so nothing reports. cohere has no warning channel for a rule; see
	// the ruling on #d21war2.
	t.Run("no pages or app directory", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, upstreamRows(nil, nil, nil, nil, nil, nil), withoutPagesDirectory...)
	})
}

// upstream's 'does not print warning ... with rootDir in context settings' row is not replayed: it
// sets `settings.next.rootDir`, and cohere carries no shared settings. The root here is the project
// root, which is the row's own default.

func TestNoHtmlLinkForPagesReadsThePagesDirectory(t *testing.T) {
	t.Parallel()

	rows := []htmlLinkRow{
		{`<a href="/">x</a>`, []string{"/"}},
		{`<a href="/about">x</a>`, []string{"/about/"}},
		{`<a href="/about/">x</a>`, []string{"/about/"}},
		// A relative href is not a route this project defines, under any page tree.
		{`<a href="about">x</a>`, nil},
		{`<a href="./about">x</a>`, nil},
		// The query, the fragment and a trailing index.html come off before matching, and the
		// message names the href as normalized.
		{`<a href="/about?x=1">x</a>`, []string{"/about/"}},
		{`<a href="/about#h">x</a>`, []string{"/about/"}},
		{`<a href="/about/index.html">x</a>`, []string{"/about/"}},
		{`<a href="/index.html">x</a>`, []string{"/"}},
		{`<a href="/blog">x</a>`, []string{"/blog/"}},
		// An index file is two routes, `/blog/` and `/blog/index`, and `[slug]` matches it too.
		{`<a href="/blog/index">x</a>`, []string{"/blog/index/", "/blog/index/"}},
		{`<a href="/blog/post">x</a>`, []string{"/blog/post/"}},
		// The dynamic wildcard refuses a dot between two characters, and crosses slashes.
		{`<a href="/blog/post.html">x</a>`, nil},
		{`<a href="/blog/a/b">x</a>`, []string{"/blog/a/b/"}},
		{`<a href="/docs/a/b">x</a>`, []string{"/docs/a/b/"}},
		{`<a href="/docs/a.b">x</a>`, nil},
		{`<a href="/docs">x</a>`, nil},
		// No file is special: `_app`, an API route and a declaration file are routes too.
		{`<a href="/_app">x</a>`, []string{"/_app/"}},
		{`<a href="/api/x">x</a>`, []string{"/api/x/"}},
		{`<a href="/a.d">x</a>`, []string{"/a.d/"}},
		// Only .js, .jsx, .ts and .tsx are pages.
		{`<a href="/readme">x</a>`, nil},
		{`<a href="/x">x</a>`, nil},
		{`<a href="/q">x</a>`, []string{"/q/"}},
		{`<a href="/r">x</a>`, []string{"/r/"}},
		{`<a href="/s">x</a>`, []string{"/s/"}},
		{`<a href="https://x.com/about">x</a>`, nil},
		{`<a href="//x/about">x</a>`, nil},
		{`<a href="">x</a>`, nil},
		{`<a href="#h">x</a>`, nil},
		{`<a href="?x=1">x</a>`, nil},
		{`<a href="/nothing">x</a>`, nil},
		{`<a href="/About">x</a>`, nil},
	}
	runHtmlLinkTree(t, rows,
		"pages/index.tsx", "pages/about.tsx", "pages/blog/index.tsx", "pages/blog/[slug].tsx", "pages/docs/[...all].tsx",
		"pages/a.d.ts", "pages/_app.tsx", "pages/api/x.ts", "pages/readme.md", "pages/x.mjs", "pages/q.jsx", "pages/r.js", "pages/s.ts")
}

func TestNoHtmlLinkForPagesReadsTheAppDirectory(t *testing.T) {
	t.Parallel()

	// Only `/` matches a static app route. The rest are app routes with no trailing slash, against
	// hrefs that always gain one, and they hold upstream's other app quirks unobservable: a nested
	// `layout` and `index` are routes, a group and a slot drop out, a top-level `layout` does not.
	t.Run("static routes", func(t *testing.T) {
		t.Parallel()
		rows := []htmlLinkRow{{`<a href="/">x</a>`, []string{"/"}}}
		for _, href := range []string{"/about", "/about/", "/about/layout", "/about/index", "/layout", "/loading", "/g", "/(group)/g", "/s", "/@slot/s", "/r", "/x/thing", "/x"} {
			rows = append(rows, htmlLinkRow{`<a href="` + href + `">x</a>`, nil})
		}
		runHtmlLinkTree(t, rows,
			"app/page.tsx", "app/layout.tsx", "app/loading.tsx", "app/about/page.tsx", "app/about/layout.tsx", "app/about/index.tsx",
			"app/(group)/g/page.tsx", "app/@slot/s/page.tsx", "app/r/route.ts", "app/x/thing.tsx")
	})

	t.Run("a dynamic route", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{
			{`<a href="/">x</a>`, []string{"/"}},
			{`<a href="/about">x</a>`, []string{"/about/"}},
			{`<a href="/a/b">x</a>`, []string{"/a/b/"}},
			{`<a href="/a.b">x</a>`, nil},
			{`<a href="/a/b.c">x</a>`, nil},
			{`<a href="about">x</a>`, nil},
		}, "app/[id]/page.tsx")
	})
}

func TestNoHtmlLinkForPagesReportsOncePerRoute(t *testing.T) {
	t.Parallel()

	t.Run("pages and app both define the root", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{{`<a href="/">x</a>`, []string{"/", "/"}}}, "app/page.tsx", "pages/index.tsx")
	})

	// The two pages directories' routes are de-duplicated together, so `/` from both is one route;
	// `src/app` adds the second.
	t.Run("root and src directories", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{
			{`<a href="/">x</a>`, []string{"/", "/"}},
			{`<a href="/only">x</a>`, []string{"/only/"}},
		}, "pages/index.tsx", "src/pages/index.tsx", "src/pages/only.tsx", "src/app/page.tsx")
	})
}

func TestNoHtmlLinkForPagesCompilesRouteNamesAsUpstreamDoes(t *testing.T) {
	t.Parallel()

	// One greedy wildcard from the first `[` to the last `]`.
	t.Run("two dynamic segments", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{
			{`<a href="/list/a/b">x</a>`, []string{"/list/a/b/"}},
			{`<a href="/list/a">x</a>`, []string{"/list/a/"}},
			{`<a href="/list/">x</a>`, nil},
			{`<a href="/z/1/mid/2">x</a>`, []string{"/z/1/mid/2/"}},
			{`<a href="/z/1/x/2">x</a>`, []string{"/z/1/x/2/"}},
			{`<a href="/z/1/2/3/mid">x</a>`, []string{"/z/1/2/3/mid/"}},
		}, "pages/list/[foo]/[id].tsx", "pages/z/[a]/mid/[b].tsx")
	})

	// A page name is a pattern, not a literal.
	t.Run("pattern characters in a page name", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{
			{`<a href="/a+b">x</a>`, nil},
			{`<a href="/aab">x</a>`, []string{"/aab/"}},
			{`<a href="/ab">x</a>`, []string{"/ab/"}},
			{`<a href="/c.d">x</a>`, []string{"/c.d/"}},
			{`<a href="/cxd">x</a>`, []string{"/cxd/"}},
			{`<a href="/e$f">x</a>`, nil},
		}, "pages/a+b.tsx", "pages/c.d.tsx", "pages/e$f.tsx")
	})

	// Divergence, kept: upstream's `new RegExp` throws on `^/c(d/$`, so ESLint fails to load the
	// rule and reports nothing in any file. Here the one route that cannot compile is dropped.
	t.Run("a page name that is not a pattern", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{{`<a href="/ok">x</a>`, []string{"/ok/"}}}, "pages/c(d.tsx", "pages/ok.tsx")
	})

	// An entry is judged by its name before its kind.
	t.Run("a directory named like a page", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{
			{`<a href="/x">x</a>`, []string{"/x/"}},
			{`<a href="/x.tsx/inner">x</a>`, nil},
		}, "pages/x.tsx/inner.tsx")
	})
}

func TestNoHtmlLinkForPagesFindsNoDirectory(t *testing.T) {
	t.Parallel()

	t.Run("no pages or app directory", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{{`<a href="/">x</a>`, nil}}, "lib/a.tsx")
	})

	// Divergence, kept: a file named `pages` passes upstream's existsSync and then readdirSync throws
	// ENOTDIR, so the rule fails to load. Here it is not a directory and is not found.
	t.Run("a file named pages", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTree(t, []htmlLinkRow{{`<a href="/">x</a>`, []string{"/"}}}, "pages", "app/page.tsx")
	})

	// A harness with no program has no disk to read, which is the no-directory case. Checked by row
	// rather than with ExpectClean, which would offer the site this source as the rule's clean
	// example, silent only because nothing was read.
	t.Run("no program", func(t *testing.T) {
		t.Parallel()
		rows := []htmlLinkRow{{`<a href="/">x</a>`, nil}}
		expectHtmlLinkRows(t, rule_testing.Run(t, NoHtmlLinkForPages, "pages/index.tsx", htmlLinkRowsSource(rows)), rows)
	})
}

// The clean examples the site shows, each beside the page it would otherwise name.
func TestNoHtmlLinkForPagesIsSilentBesideAPage(t *testing.T) {
	t.Parallel()

	for _, element := range []string{
		`<a href="/about" target="_blank">About</a>`,
		`<a href="/contact">Contact</a>`,
		`<Link href="/about">About</Link>`,
	} {
		t.Run(element, func(t *testing.T) {
			t.Parallel()
			source := "export const C = () => (" + element + ");"
			result := rule_testing.RunTypedFiles(t, NoHtmlLinkForPages, map[string]string{"foo.tsx": source, "pages/about.tsx": "export {};"}, "foo.tsx")
			rule_testing.ExpectClean(t, result)
		})
	}
}

func TestNoHtmlLinkForPagesReadsTheAnchorAsUpstreamDoes(t *testing.T) {
	t.Parallel()

	runHtmlLinkTree(t, []htmlLinkRow{
		{`<a href="/about">x</a>`, []string{"/about/"}},
		{`<a href="/about" />`, []string{"/about/"}},
		{`<Link href="/about">x</Link>`, nil},
		{`<A href="/about">x</A>`, nil},
		{`<a.b href="/about">x</a.b>`, nil},
		{`<a>x</a>`, nil},
		// Only a plain string `_blank` exempts, and only on an attribute spelled `target`.
		{`<a target="_blank" href="/about">x</a>`, nil},
		{`<a href="/about" target={"_blank"}>x</a>`, []string{"/about/"}},
		{`<a href="/about" target="_self">x</a>`, []string{"/about/"}},
		{`<a href="/about" TARGET="_blank">x</a>`, []string{"/about/"}},
		// Only a plain string href is read, and the first one.
		{`<a href={"/about"}>x</a>`, nil},
		{`<a HREF="/about">x</a>`, nil},
		{`<a href="/about" href="/nothing">x</a>`, []string{"/about/"}},
		{`<a href="/nothing" href="/about">x</a>`, nil},
		// download exempts by presence, whatever its value.
		{`<a href="/about" download>x</a>`, nil},
		{`<a href="/about" download={false}>x</a>`, nil},
		{`<a {...p} href="/about">x</a>`, []string{"/about/"}},
		// Divergence, kept: upstream throws reading `.value.value` of an attribute with no value,
		// and ESLint reports nothing for the file. A bare href has no route to match; a bare target
		// is not `_blank`.
		{`<a href>x</a>`, nil},
		{`<a href="/about" target>x</a>`, []string{"/about/"}},
		// ESLint's parser decodes HTML entities in a JSX attribute string, and so does
		// jsx.StringAttributeValue, so an encoded slash is the route it spells (#51y9jh2).
		{`<a href="&#47;about">x</a>`, []string{"/about/"}},
		{`<a href="&#47;nothing">x</a>`, nil},
	}, "pages/about.tsx")
}

func TestNoHtmlLinkForPagesFollowsSymbolicLinksAsNodeLists(t *testing.T) {
	t.Parallel()

	// A link to a page file is a page, judged by its name. A link to a directory is not walked,
	// because Node's dirent reports the link and not the directory.
	rows := []htmlLinkRow{
		{`<a href="/link">x</a>`, []string{"/link/"}},
		{`<a href="/dirlink/deep">x</a>`, nil},
	}
	result := rule_testing.RunTypedFilesWithSetup(t, NoHtmlLinkForPages, htmlLinkTree(rows, "real/linked.tsx", "pages/index.tsx", "realdir/deep.tsx"), "foo.tsx", func(directory string) {
		if err := os.Symlink(filepath.Join("..", "real", "linked.tsx"), filepath.Join(directory, "pages", "link.tsx")); err != nil {
			t.Fatalf("linking the page file: %v", err)
		}
		if err := os.Symlink(filepath.Join("..", "realdir"), filepath.Join(directory, "pages", "dirlink")); err != nil {
			t.Fatalf("linking the directory: %v", err)
		}
	})
	expectHtmlLinkRows(t, result, rows)
}

func TestNoHtmlLinkForPagesTakesPagesDirectories(t *testing.T) {
	t.Parallel()

	// An empty list is truthy upstream, so it names no pages directory, and the app directory still
	// counts.
	t.Run("an empty list with an app directory", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTreeWithOption(t, []htmlLinkRow{{`<a href="/">x</a>`, []string{"/"}}}, `[]`, "pages/index.tsx", "app/page.tsx")
	})
	t.Run("an empty list alone", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTreeWithOption(t, []htmlLinkRow{{`<a href="/">x</a>`, nil}}, `[]`, "pages/index.tsx")
	})

	// An empty string is falsy upstream, so it is the default.
	t.Run("an empty string", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTreeWithOption(t, []htmlLinkRow{{`<a href="/">x</a>`, []string{"/"}}}, `""`, "pages/index.tsx")
	})

	t.Run("a directory that does not exist", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTreeWithOption(t, []htmlLinkRow{{`<a href="/">x</a>`, nil}}, `"nope"`, "pages/index.tsx")
	})

	// The option replaces the pages directories and never the app directories.
	t.Run("a custom directory beside pages and app", func(t *testing.T) {
		t.Parallel()
		runHtmlLinkTreeWithOption(t, []htmlLinkRow{
			{`<a href="/">x</a>`, []string{"/"}},
			{`<a href="/about">x</a>`, []string{"/about/"}},
		}, `"custom"`, "custom/about.tsx", "pages/index.tsx", "app/page.tsx")
	})
}

// Divergence, kept and ruled on #d21war2: a relative pages directory resolves against the config
// file's directory, where upstream resolves it against the process's working directory. Pinned with
// the config one level below the root, the only place the two readings part: `custom` there is
// `sub/custom`, where ESLint started at the root would read the root's `custom`.
func TestNoHtmlLinkForPagesAnchorsRelativeDirectoriesToTheConfig(t *testing.T) {
	t.Parallel()

	rows := []htmlLinkRow{
		{`<a href="/sub-page">x</a>`, []string{"/sub-page/"}},
		{`<a href="/root-page">x</a>`, nil},
	}
	files := htmlLinkTree(rows, "sub/custom/sub-page.tsx", "custom/root-page.tsx")
	result := rule_testing.RunTypedFilesWithOptionsFor(t, NoHtmlLinkForPages, files, "foo.tsx", func(directory string) any {
		decoded, err := decodeNoHtmlLinkForPagesOptions([]byte(`"custom"`), rule.OptionsBase{ConfigDirectory: filepath.Join(directory, "sub"), ProjectRoot: directory})
		if err != nil {
			t.Fatalf("decoding: %v", err)
		}
		return decoded
	})
	expectHtmlLinkRows(t, result, rows)
}

func TestNoHtmlLinkForPagesDecodesUpstreamsSchema(t *testing.T) {
	t.Parallel()

	base := rule.OptionsBase{ConfigDirectory: "/project/config", ProjectRoot: "/project"}
	accepted := []struct {
		raw  string
		want NoHtmlLinkForPagesOptions
	}{
		{``, NoHtmlLinkForPagesOptions{}},
		{`""`, NoHtmlLinkForPagesOptions{}},
		{`"pages"`, NoHtmlLinkForPagesOptions{Configured: true, PagesDirectories: []string{"/project/config/pages"}}},
		{`"/elsewhere/pages"`, NoHtmlLinkForPagesOptions{Configured: true, PagesDirectories: []string{"/elsewhere/pages"}}},
		{`[]`, NoHtmlLinkForPagesOptions{Configured: true, PagesDirectories: []string{}}},
		{`["a", "../b", ""]`, NoHtmlLinkForPagesOptions{Configured: true, PagesDirectories: []string{"/project/config/a", "/project/b", ""}}},
	}
	for _, testCase := range accepted {
		t.Run("accepts "+testCase.raw, func(t *testing.T) {
			t.Parallel()
			var raw []byte
			if testCase.raw != "" {
				raw = []byte(testCase.raw)
			}
			decoded, err := decodeNoHtmlLinkForPagesOptions(raw, base)
			if err != nil {
				t.Fatalf("refused: %v", err)
			}
			got := decoded.(NoHtmlLinkForPagesOptions)
			if got.Configured != testCase.want.Configured || !slices.Equal(got.PagesDirectories, testCase.want.PagesDirectories) {
				t.Errorf("decoded %+v, want %+v", got, testCase.want)
			}
		})
	}

	refused := []struct {
		raw  string
		base rule.OptionsBase
		want string
	}{
		{`{}`, base, "expected a pages directory or a list of them"},
		{`3`, base, "expected a pages directory or a list of them"},
		{`true`, base, "expected a pages directory or a list of them"},
		{`["a", 3]`, base, "element 2 of the pages directory list is 3"},
		{`["a", null]`, base, "element 2 of the pages directory list is null"},
		{`["a", "a"]`, base, `names "a" twice`},
		{`"pages"`, rule.OptionsBase{}, "is relative and the config's directory is not known"},
	}
	for _, testCase := range refused {
		t.Run("refuses "+testCase.raw, func(t *testing.T) {
			t.Parallel()
			_, err := decodeNoHtmlLinkForPagesOptions([]byte(testCase.raw), testCase.base)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Errorf("error %v, want one containing %q", err, testCase.want)
			}
		})
	}
}

// The finding points at the opening element, which is what upstream reports: column 25 to the end
// of `<a href='/about'>` in a measured row, and the whole element when it closes itself.
func TestNoHtmlLinkForPagesPointsAtTheOpeningElement(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct{ element, reported string }{
		{`<a href='/about'>About</a>`, `<a href='/about'>`},
		{`<a href='/about' />`, `<a href='/about' />`},
	} {
		t.Run(testCase.element, func(t *testing.T) {
			t.Parallel()
			source := "export const C = () => (" + testCase.element + ");"
			result := rule_testing.RunTypedFiles(t, NoHtmlLinkForPages, map[string]string{"foo.tsx": source, "pages/about.tsx": "export {};"}, "foo.tsx")
			rule_testing.ExpectFindings(t, result, messageNoHtmlLinkForPages.Id)
			diagnostic := result.Diagnostics[0]
			if reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]; reported != testCase.reported {
				t.Errorf("finding points at %q, want %q", reported, testCase.reported)
			}
		})
	}
}

// htmlLinkFingerprintOf is the rule's program fingerprint over a fixture tree on disk, read through the
// Program the walk hands the rule, under option written as runHtmlLinkTreeWithOption writes it, or under
// no option at all when it is "".
func htmlLinkFingerprintOf(t *testing.T, files map[string]string, option string) [sha256.Size]byte {
	t.Helper()
	var fingerprint [sha256.Size]byte
	probe := NoHtmlLinkForPages
	probe.Run = func(ctx rule.Context, options any) rule.Listeners {
		fingerprint = NoHtmlLinkForPages.ProgramFingerprint(ctx.Program, options)
		return nil
	}
	rule_testing.RunTypedFilesWithOptionsFor(t, probe, files, "foo.tsx", func(directory string) any {
		if option == "" {
			return nil
		}
		quoted, _ := json.Marshal(directory)
		raw := strings.ReplaceAll(option, "{root}", strings.Trim(string(quoted), `"`))
		decoded, err := decodeNoHtmlLinkForPagesOptions([]byte(raw), rule.OptionsBase{ConfigDirectory: directory, ProjectRoot: directory})
		if err != nil {
			t.Fatalf("decoding %s: %v", raw, err)
		}
		return decoded
	})
	if fingerprint == ([sha256.Size]byte{}) {
		t.Fatal("the probe never ran, so no fingerprint was read")
	}
	return fingerprint
}

// htmlLinkFingerprintCase is one change to a fixture tree: paths to add as empty pages, paths to remove,
// and files to write with the text given.
type htmlLinkFingerprintCase struct {
	add    []string
	remove []string
	write  map[string]string
}

// expectHtmlLinkFingerprint checks each case against the base tree's fingerprint: every case in moves must
// change it and every case in holds must leave it.
func expectHtmlLinkFingerprint(t *testing.T, base []string, option string, moves, holds map[string]htmlLinkFingerprintCase) {
	t.Helper()
	tree := func(change htmlLinkFingerprintCase) map[string]string {
		files := htmlLinkTree([]htmlLinkRow{{`<a href="/about">x</a>`, nil}}, base...)
		for _, path := range change.add {
			files[path] = "export {};"
		}
		for _, path := range change.remove {
			delete(files, path)
		}
		for path, text := range change.write {
			files[path] = text
		}
		return files
	}
	original := htmlLinkFingerprintOf(t, tree(htmlLinkFingerprintCase{}), option)
	for name, change := range moves {
		if htmlLinkFingerprintOf(t, tree(change), option) == original {
			t.Errorf("%s left the fingerprint unchanged, so a stale finding would replay", name)
		}
	}
	for name, change := range holds {
		if htmlLinkFingerprintOf(t, tree(change), option) != original {
			t.Errorf("%s moved the fingerprint, so files whose records cannot change were walked again", name)
		}
	}
}

// The fingerprint is the whole of what a file's record reads beyond the file (#s9k38p3), proven both ways:
// every change to the routes moves it, a route found a second way included, since it reports a second
// time, and so does a directory appearing, since that is the difference between listening and not; and
// every change the routes cannot see leaves it, so the findings cache replays across it.
func TestNoHtmlLinkForPagesFingerprintsExactlyTheRoutes(t *testing.T) {
	t.Parallel()

	base := []string{"pages/index.tsx", "pages/about.tsx", "pages/blog/[slug].tsx", "app/shop/[id]/page.tsx", "lib/util.ts"}
	expectHtmlLinkFingerprint(t, base, "", map[string]htmlLinkFingerprintCase{
		"a page added":                     {add: []string{"pages/contact.tsx"}},
		"a page removed":                   {remove: []string{"pages/about.tsx"}},
		"a page renamed":                   {add: []string{"pages/team.tsx"}, remove: []string{"pages/about.tsx"}},
		"a page moved into a directory":    {add: []string{"pages/info/about.tsx"}, remove: []string{"pages/about.tsx"}},
		"an index page added below":        {add: []string{"pages/blog/index.tsx"}},
		"an app page added":                {add: []string{"app/cart/[id]/page.tsx"}},
		"the root defined by app as well":  {add: []string{"app/page.tsx"}},
		"a src pages directory beside":     {add: []string{"src/pages/only.tsx"}},
		"a page under a dynamic directory": {add: []string{"pages/blog/[slug]/comments.tsx"}},
		// Routes are de-duplicated on their text before the wildcard goes in, so `/shop/[a]/[id]` and
		// `/shop/[id]` are two routes that compile alike, and a matching anchor reports twice.
		"a second app route the wildcard compiles alike": {add: []string{"app/shop/[a]/[id]/page.tsx"}},
	}, map[string]htmlLinkFingerprintCase{
		"a page's contents edited":           {write: map[string]string{"pages/about.tsx": "export const about = 1;"}},
		"the file checked edited":            {write: map[string]string{"foo.tsx": htmlLinkRowsSource([]htmlLinkRow{{`<a href="/nothing">x</a>`, nil}})}},
		"a file outside the directories":     {add: []string{"lib/more.ts", "components/about.tsx"}},
		"a file in pages that is not a page": {add: []string{"pages/readme.md", "pages/styles.css", "pages/x.mjs"}},
		"a dynamic segment renamed":          {add: []string{"pages/blog/[id].tsx"}, remove: []string{"pages/blog/[slug].tsx"}},
		"a top-level app layout":             {add: []string{"app/layout.tsx"}},
		"a page whose name does not compile": {add: []string{"pages/c(d.tsx"}},
	})

	// With no pages or app directory the rule attaches no listener; with an empty one it listens and
	// reports nothing. The cache replays coverage with findings, so the first page-less directory moves it,
	// and the routes alone could not tell.
	expectHtmlLinkFingerprint(t, []string{"lib/util.ts"}, "", map[string]htmlLinkFingerprintCase{
		"a pages directory with no page in it": {add: []string{"pages/readme.md"}},
		"the first page":                       {add: []string{"pages/index.tsx"}},
	}, map[string]htmlLinkFingerprintCase{
		"a file outside the directories": {add: []string{"lib/more.ts"}},
	})
}

// Under a configured pages directory the fingerprint is that directory's routes (#s9k38p3): a page added
// there moves it, and a page added to the default `pages/`, which the setting replaced, does not.
func TestNoHtmlLinkForPagesFingerprintsTheConfiguredPagesDirectory(t *testing.T) {
	t.Parallel()

	base := []string{"custom/about.tsx", "pages/index.tsx", "lib/util.ts"}
	expectHtmlLinkFingerprint(t, base, `"custom"`, map[string]htmlLinkFingerprintCase{
		"a page added under it":   {add: []string{"custom/team.tsx"}},
		"a page removed under it": {remove: []string{"custom/about.tsx"}},
		"an app page added":       {add: []string{"app/page.tsx"}},
	}, map[string]htmlLinkFingerprintCase{
		"a page added to the pages directory it replaced": {add: []string{"pages/team.tsx"}},
		"a page removed from the pages directory":         {remove: []string{"pages/index.tsx"}},
	})

	// One tree, two settings: the option chooses the data, so the fingerprint follows it.
	files := htmlLinkTree(nil, "custom/about.tsx", "other/team.tsx", "pages/index.tsx")
	fingerprints := map[[sha256.Size]byte]string{}
	for _, option := range []string{"", `"custom"`, `"other"`, `["custom", "other"]`} {
		fingerprint := htmlLinkFingerprintOf(t, files, option)
		if earlier, seen := fingerprints[fingerprint]; seen {
			t.Errorf("the settings %q and %q fingerprint one tree alike, though they read different pages", earlier, option)
		}
		fingerprints[fingerprint] = option
	}
}

// countingProgram is a Program whose file system counts the existence checks asked of it, by path.
type countingProgram struct {
	rule.Program
	fileSystem *countingFileSystem
}

func (program countingProgram) FS() vfs.FS { return program.fileSystem }

// countingFileSystem counts DirectoryExists calls by path. A route model build asks each candidate
// directory once, so the count for one of them is how many builds there were.
type countingFileSystem struct {
	vfs.FS
	mutex  sync.Mutex
	checks map[string]int
}

func (fileSystem *countingFileSystem) DirectoryExists(path tspath.RootedDirectoryPath) bool {
	fileSystem.mutex.Lock()
	fileSystem.checks[path.AsString()]++
	fileSystem.mutex.Unlock()
	return fileSystem.FS.DirectoryExists(path)
}

// Not parallel: the route models are one process-wide slot keyed on the program, and a parallel test
// building another program's model would empty it between these calls and count a second build that
// was not this test's.
//
// The walk takes a rule's program fingerprint on every worker that serves a selection, so several
// workers ask at once (#s9k38p3). They must all get one answer from one build of the route model, which
// reads the disk: sixteen goroutines asking together build it once.
func TestNoHtmlLinkForPagesFingerprintsOnSeveralWorkersFromOneBuild(t *testing.T) {
	const workers = 16
	var fingerprints [workers][sha256.Size]byte
	var fileSystem *countingFileSystem
	probe := NoHtmlLinkForPages
	probe.Run = func(ctx rule.Context, options any) rule.Listeners {
		fileSystem = &countingFileSystem{FS: ctx.Program.FS(), checks: map[string]int{}}
		program := countingProgram{Program: ctx.Program, fileSystem: fileSystem}
		var wait sync.WaitGroup
		for worker := range workers {
			wait.Add(1)
			go func() {
				defer wait.Done()
				fingerprints[worker] = NoHtmlLinkForPages.ProgramFingerprint(program, nil)
			}()
		}
		wait.Wait()
		return nil
	}
	rule_testing.RunTypedFiles(t, probe, htmlLinkTree(nil, "pages/index.tsx", "pages/about.tsx", "app/page.tsx"), "foo.tsx")

	if fileSystem == nil {
		t.Fatal("the probe never ran, so nothing was asked")
	}
	builds := 0
	for path, count := range fileSystem.checks {
		if filepath.Base(path) == "pages" && filepath.Base(filepath.Dir(path)) != "src" {
			builds = count
		}
	}
	if builds != 1 {
		t.Errorf("%d workers asking at once built the route model %d times, want once (existence checks %v)", workers, builds, fileSystem.checks)
	}
	for worker := 1; worker < workers; worker++ {
		if fingerprints[worker] != fingerprints[0] {
			t.Fatalf("worker %d got a different fingerprint from worker 0 for one program", worker)
		}
	}
	if fingerprints[0] == ([sha256.Size]byte{}) {
		t.Fatal("no fingerprint was read")
	}
}

// The message is asserted whole, with the href interpolated as upstream interpolates it, normalized.
func TestNoHtmlLinkForPagesNamesTheRouteInItsMessage(t *testing.T) {
	t.Parallel()

	want := "This is a raw <a> element navigating to `/about/`, a route this project's pages or app " +
		"directory defines. Use `Link` from `next/link`, which navigates on the client and prefetches " +
		"the destination. A plain <a> does a full document load instead, so the application is torn " +
		"down and rebuilt and every piece of in-memory state goes with it."
	if got := noHtmlLinkForPagesMessage("/about/").Description; got != want {
		t.Errorf("description is %q, want %q", got, want)
	}
}
