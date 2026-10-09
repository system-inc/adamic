CODE UNDER TEST: Adamic internal/load wrappers, configuration, declarations, diagnostics, source filesystem and library selection. Upstream checker and installed Node declarations are not mutated.
ORACLE: hand-written acceptance, diagnostic and declaration assertions; TypeScript diagnostic authority, official Node type signatures, and Adamic docs/0.1.md contract. No scoped row runs an external comparator.
Functions available on the reached paths (read-only FS methods are interface capabilities; not every invocation was dynamically traced):
{
  "internal/load/load.go": [
    "func (e *CheckError) Error() string {",
    "func compilerOptions() *core.CompilerOptions {",
    "func Load(paths []string) (*Program, error) {",
    "func LoadOverlay(paths []string, overlay map[string]string) (*Program, error) {",
    "func load(paths []string, overlay map[string]string) (*Program, error) {",
    "func (p *Program) Files() []*ast.SourceFile {",
    "func (p *Program) Checker(ctx context.Context, sourceFile *ast.SourceFile) (*checker.Checker, func()) {",
    "func (p *Program) FileName(sourceFile *ast.SourceFile) string {",
    "func (p *Program) Where(node *ast.Node) string {",
    "func IsPrelude(sourceFile *ast.SourceFile) bool {",
    "func IsLibrary(sourceFile *ast.SourceFile) bool {",
    "func rootFileName(fs *sourceFS, currentDirectory tspath.RootedDirectoryPath, path string) (tspath.RootedFilePath, error) {",
    "func (p *Program) diagnostics(ctx context.Context) []string {",
    "func (p *Program) formatDiagnostic(diagnostic *ast.Diagnostic) string {",
    "func writeChain(builder *strings.Builder, chain []*ast.Diagnostic, depth int) {",
    "func (p *Program) lineAndColumn(sourceFile *ast.SourceFile, position int) (int, int) {",
    "func (p *Program) CompilerProgram() *compiler.Program { return p.compiler }"
  ],
  "internal/load/export_collision.go": [
    "func starCollision(diagnostic *ast.Diagnostic) string {"
  ],
  "internal/load/source_fs.go": [
    "func (s *sourceFS) adamicFile(path tspath.RootedFilePath) (tspath.RootedFilePath, bool) {",
    "func (s *sourceFS) displayName(path tspath.RootedFilePath) string {",
    "func (s *sourceFS) FileExists(path tspath.RootedFilePath) bool {",
    "func (s *sourceFS) ReadFile(path tspath.RootedFilePath) (string, bool) {",
    "func (s *sourceFS) DirectoryExists(path tspath.RootedDirectoryPath) bool {",
    "func (s *sourceFS) Stat(path tspath.RootedPath) vfs.FileInfo {",
    "func (s *sourceFS) Realpath(path tspath.RootedPath) tspath.RootedPath {",
    "func (s *sourceFS) WriteFile(path tspath.RootedFilePath, data string) error {",
    "func (s *sourceFS) AppendFile(path tspath.RootedFilePath, data string) error {",
    "func (s *sourceFS) Remove(path tspath.RootedPath) error {",
    "func (s *sourceFS) Chtimes(path tspath.RootedPath, aTime time.Time, mTime time.Time) error {"
  ],
  "internal/load/node_library.go": [
    "func nodeTypesIndex(directory string) (string, error) {",
    "func usesNodeModules(program *compiler.Program) bool {",
    "func IsNodeLibrary(file *ast.SourceFile) bool {",
    "func nodePrelude() string {"
  ],
  "internal/load/declarations.go": [
    "func (p *Program) Declarations(ctx context.Context) []Declaration {",
    "func (p *Program) declaration(typeChecker *checker.Checker, sourceFile *ast.SourceFile, node *ast.Node, kind string, name *ast.Node) Declaration {"
  ],
  "internal/load/regexp_library.go": [
    "func (s *regexpLibraryFS) ReadFile(path tspath.RootedFilePath) (string, bool) {"
  ]
}
Fixed menu was declared before inspecting mutant failures. M06 changes the default-false NoImplicitReturns option to true. M13 drops the whole loop. M20 may be equivalent because every .d.ts also ends in .ts.
