package next

import (
	"crypto/sha256"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// The route model is @next/eslint-plugin-next 16.3.1's, from `dist/rules/no-html-link-for-pages.js`
// and `dist/utils/url.js`, ported function for function so the two read side by side. Every quirk
// below is upstream's, measured against the installed plugin rather than read off its source, and kept
// because this rule answers to upstream's name and its ESLint twin runs upstream's code.
//
// The quirks a reader is most likely to "fix", each pinned by a fixture:
//
//   - A pages index file pushes two routes, `dir/` and `dir/index`, so `/blog/index` is a route too.
//   - Below the top of an app directory the walk uses the PAGES parser, so a nested `layout.tsx`
//     becomes the route `/about/layout` and a nested `index.tsx` pushes two, as in pages.
//   - App routes are anchored with no trailing slash while every href is normalized with one, so a
//     static app route other than `/` never matches. Only `/` and routes with a dynamic segment do.
//   - Route names are compiled as regular expressions without escaping, so `a+b.tsx` matches `/aab`
//     and not `/a+b`, and `c.d.tsx` matches `/cxd`.
//   - A dynamic segment is replaced by one greedy match from the first `[` to the last `]`, so
//     `/z/[a]/mid/[b]` becomes one wildcard and `/z/1/2/3/mid` matches it.
//   - A route found in more than one way reports once per way: `/` in both `pages/` and `app/`
//     reports twice on one anchor.

// noHtmlLinkForPagesRouteModel is the routes one root's directories define, compiled.
type noHtmlLinkForPagesRouteModel struct {
	// found is whether any pages or app directory exists. Without one upstream returns an empty
	// visitor and the rule reports nothing at all.
	found bool

	// routes are upstream's `allUrlRegex`: the pages routes, then the app routes, each list
	// de-duplicated on its own source text.
	routes []*esregexp.RegExp
}

// noHtmlLinkForPagesRouteModels holds the models built for one program, so each is built once per
// run rather than once per file.
//
// Within a run the directories are read once, the way upstream's process-wide existsSync and
// readdirSync caches read them once per ESLint process, and the fingerprint the findings cache keys
// the rule on is read off the same model, so a page added between runs moves it and cannot replay a
// stale verdict. Keyed on the program's identity, so the next run's program starts a fresh set rather
// than reading this one's, the same shape as the Tailwind design system's cache.
var noHtmlLinkForPagesRouteModels struct {
	sync.Mutex
	program rule.ProgramIdentity
	models  map[string]*noHtmlLinkForPagesRouteModel
}

// noHtmlLinkForPagesRouteModelFor returns the route model for one root and one setting of the option,
// building it on the first ask in this program.
func noHtmlLinkForPagesRouteModelFor(program rule.Program, root string, settings NoHtmlLinkForPagesOptions) *noHtmlLinkForPagesRouteModel {
	key := root + "\x00" + noHtmlLinkForPagesSettingsKey(settings)

	noHtmlLinkForPagesRouteModels.Lock()
	defer noHtmlLinkForPagesRouteModels.Unlock()

	if noHtmlLinkForPagesRouteModels.program != program.Identity() || noHtmlLinkForPagesRouteModels.models == nil {
		noHtmlLinkForPagesRouteModels.program = program.Identity()
		noHtmlLinkForPagesRouteModels.models = map[string]*noHtmlLinkForPagesRouteModel{}
	}
	if model, built := noHtmlLinkForPagesRouteModels.models[key]; built {
		return model
	}
	model := buildNoHtmlLinkForPagesRouteModel(program.FS(), root, settings)
	noHtmlLinkForPagesRouteModels.models[key] = model
	return model
}

// noHtmlLinkForPagesFingerprint is the rule's program fingerprint under one setting of its option, which
// chooses the pages directories: whether a pages or app directory was found, and every route the model
// holds, as compiled, in the order a file's anchors are tested against them (#s9k38p3).
//
// A file's record reads its own bytes and these two, nothing else. The routes decide its findings: an
// anchor reports once for each route it matches, and the message names only the href, so duplicates are
// kept, since a route found two ways reports twice. Whether a directory was found decides its coverage,
// which the cache replays with the findings: with none the rule attaches no listener, and with an empty
// one it listens and reports nothing. Two things the model is built from are left out because no record
// can tell them apart. A route that does not compile is dropped before any anchor sees it. The root
// reaches a record only through the routes, and is in the findings key besides. A page's contents are
// never read, so an edit inside one leaves the fingerprint alone.
func noHtmlLinkForPagesFingerprint(program rule.Program, options any) [sha256.Size]byte {
	settings, _ := rule.OptionsAs[NoHtmlLinkForPagesOptions](options)
	model := noHtmlLinkForPagesRouteModelFor(program, program.GetCurrentDirectory(), settings)
	hash := sha256.New()
	if model.found {
		hash.Write([]byte{1})
	}
	for _, route := range model.routes {
		hash.Write([]byte(route.Source()))
		hash.Write([]byte{0})
	}
	var fingerprint [sha256.Size]byte
	copy(fingerprint[:], hash.Sum(nil))
	return fingerprint
}

// noHtmlLinkForPagesSettingsKey spells an option setting as a cache key that no other setting shares.
func noHtmlLinkForPagesSettingsKey(settings NoHtmlLinkForPagesOptions) string {
	if !settings.Configured {
		return "default"
	}
	return "configured\x00" + strings.Join(settings.PagesDirectories, "\x00")
}

// buildNoHtmlLinkForPagesRouteModel is upstream's `create` up to the visitor: find the directories,
// walk them, compile the routes.
//
// One reading differs, on input upstream crashes on. Upstream asks `fs.existsSync`, which is true for
// a file as well as a directory, and then `readdirSync` throws ENOTDIR and the rule fails to load. A
// file named `pages` is not a pages directory, so here it is simply not found.
func buildNoHtmlLinkForPagesRouteModel(fileSystem vfs.FS, root string, settings NoHtmlLinkForPagesOptions) *noHtmlLinkForPagesRouteModel {
	pagesDirectories := settings.PagesDirectories
	if !settings.Configured {
		pagesDirectories = []string{filepath.Join(root, "pages"), filepath.Join(root, "src", "pages")}
	}
	appDirectories := []string{filepath.Join(root, "app"), filepath.Join(root, "src", "app")}

	foundPagesDirectories := existingDirectories(fileSystem, pagesDirectories)
	foundAppDirectories := existingDirectories(fileSystem, appDirectories)
	if len(foundPagesDirectories) == 0 && len(foundAppDirectories) == 0 {
		return &noHtmlLinkForPagesRouteModel{}
	}

	var pageUrls []string
	for _, directory := range foundPagesDirectories {
		for _, url := range parseUrlForPages(fileSystem, "/", directory) {
			normalized, _ := normalizeURL(url)
			pageUrls = append(pageUrls, "^"+normalized+"$")
		}
	}
	var appUrls []string
	for _, directory := range foundAppDirectories {
		for _, url := range parseUrlForAppDirectory(fileSystem, "/", directory) {
			appUrls = append(appUrls, "^"+normalizeAppPath(url)+"$")
		}
	}

	return &noHtmlLinkForPagesRouteModel{
		found:  true,
		routes: append(compileRouteSources(pageUrls), compileRouteSources(appUrls)...),
	}
}

// existingDirectories is upstream's `filter(fs.existsSync)`, asked of directories only.
func existingDirectories(fileSystem vfs.FS, directories []string) []string {
	found := make([]string, 0, len(directories))
	for _, directory := range directories {
		if directory != "" && fileSystem.DirectoryExists(tspath.RootedDirectoryPath(directory)) {
			found = append(found, directory)
		}
	}
	return found
}

// pageFileExtension and the two leaf names are upstream's patterns, which its own TODO says should
// account for every page extension and do not.
var (
	pageFileExtension = regexp.MustCompile(`\.[jt]sx?$`)
	indexPageFile     = regexp.MustCompile(`^index\.[jt]sx?$`)
	pagePageFile      = regexp.MustCompile(`^page\.[jt]sx?$`)
	layoutPageFile    = regexp.MustCompile(`^layout\.[jt]sx?$`)
)

// parseUrlForPages is upstream's `parseUrlForPages`: every page file under a directory as a route.
//
// An entry is judged by its name before its kind, as upstream judges a dirent, so a directory named
// `x.tsx` is a page and a symbolic link to a page file is a page. A directory is walked only when it
// is not a symbolic link, because Node's dirent for a link says it is a link and not a directory. A
// link whose target is missing is left out of the listing here and kept by Node, where it would push
// a route for a page that cannot load; that is the one listing difference.
func parseUrlForPages(fileSystem vfs.FS, urlPrefix string, directory string) []string {
	entries := fileSystem.GetAccessibleEntries(tspath.RootedDirectoryPath(directory))
	var urls []string
	visit := func(name string, isDirectory bool) {
		if pageFileExtension.MatchString(name) {
			if indexPageFile.MatchString(name) {
				urls = append(urls, urlPrefix)
			}
			urls = append(urls, urlPrefix+pageFileExtension.ReplaceAllLiteralString(name, ""))
			return
		}
		if _, isLink := entries.Symlinks[name]; isDirectory && !isLink {
			urls = append(urls, parseUrlForPages(fileSystem, urlPrefix+name+"/", filepath.Join(directory, name))...)
		}
	}
	for _, name := range entries.Files {
		visit(name, false)
	}
	for _, name := range entries.Directories {
		visit(name, true)
	}
	return urls
}

// parseUrlForAppDirectory is upstream's `parseUrlForAppDir`: the top of an app directory, where
// `page` is the directory's own route and `layout` is no route, and below which every directory is
// walked with parseUrlForPages, as upstream walks it.
func parseUrlForAppDirectory(fileSystem vfs.FS, urlPrefix string, directory string) []string {
	entries := fileSystem.GetAccessibleEntries(tspath.RootedDirectoryPath(directory))
	var urls []string
	visit := func(name string, isDirectory bool) {
		if pageFileExtension.MatchString(name) {
			switch {
			case pagePageFile.MatchString(name):
				urls = append(urls, urlPrefix)
			case !layoutPageFile.MatchString(name):
				urls = append(urls, urlPrefix+pageFileExtension.ReplaceAllLiteralString(name, ""))
			}
			return
		}
		if _, isLink := entries.Symlinks[name]; isDirectory && !isLink {
			urls = append(urls, parseUrlForPages(fileSystem, urlPrefix+name+"/", filepath.Join(directory, name))...)
		}
	}
	for _, name := range entries.Files {
		visit(name, false)
	}
	for _, name := range entries.Directories {
		visit(name, true)
	}
	return urls
}

// normalizeURL is upstream's `normalizeURL`. The second return is false where upstream returns
// undefined, for an empty url.
func normalizeURL(url string) (string, bool) {
	if url == "" {
		return "", false
	}
	url, _, _ = strings.Cut(url, "?")
	url, _, _ = strings.Cut(url, "#")
	if strings.HasSuffix(url, "/index.html") {
		url = strings.TrimSuffix(url, "index.html")
	}
	// An href that was only a query or a fragment stays empty rather than becoming `/`.
	if url == "" {
		return url, true
	}
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url, true
}

// normalizeAppPath is upstream's `normalizeAppPath`: groups, parallel slots and a trailing `page` or
// `route` leaf drop out of the route, and it gains a leading slash and no trailing one.
func normalizeAppPath(route string) string {
	segments := strings.Split(route, "/")
	var pathname strings.Builder
	for index, segment := range segments {
		switch {
		case segment == "":
		case segment[0] == '(' && strings.HasSuffix(segment, ")"):
		case segment[0] == '@':
		case (segment == "page" || segment == "route") && index == len(segments)-1:
		default:
			pathname.WriteString("/")
			pathname.WriteString(segment)
		}
	}
	if pathname.Len() == 0 {
		return "/"
	}
	return pathname.String()
}

// dynamicSegments is upstream's `/\[.*\]/g`. JavaScript's `.` stops at four line terminators where
// Go's stops at one, so they are spelled out. Leftmost and greedy in both engines, so one match runs
// from the first `[` to the last `]` and every dynamic segment between them becomes one wildcard.
var dynamicSegments = regexp.MustCompile("\\[[^\n\r\u2028\u2029]*\\]")

// dynamicSegmentPattern is upstream's replacement: any text with no dot between two characters.
const dynamicSegmentPattern = `((?!.+?\..+?).*?)`

// compileRouteSources is upstream's tail of `getUrlFrom...`: de-duplicate the anchored sources, then
// widen dynamic segments and compile each.
//
// A source that does not compile is left out. Upstream compiles with `new RegExp`, which throws on a
// route like `c(d`, and the rule then fails to load for every file. One page with a parenthesis in its
// name is no reason to report nothing anywhere, so here that one route is dropped and the rest stand.
func compileRouteSources(sources []string) []*esregexp.RegExp {
	seen := make(map[string]bool, len(sources))
	routes := make([]*esregexp.RegExp, 0, len(sources))
	for _, source := range sources {
		if seen[source] {
			continue
		}
		seen[source] = true
		compiled, err := esregexp.Compile(dynamicSegments.ReplaceAllLiteralString(source, dynamicSegmentPattern), "")
		if err != nil {
			continue
		}
		routes = append(routes, compiled)
	}
	return routes
}
