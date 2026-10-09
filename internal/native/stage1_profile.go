package native

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strings"
)

// Stage1ProfileUnit records both the source identity seen by clang and its bytes.
// Build currently emits one unit. A future split must enumerate every unit here.
type Stage1ProfileUnit struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

// Stage1ProfileManifest binds a text profile to the bytes the compiler consumes.
// Toolchain and target checks are conservative: another host regenerates its own profile.
type Stage1ProfileManifest struct {
	Version           int                 `json:"version"`
	Units             []Stage1ProfileUnit `json:"emitted_units"`
	CompilerSHA256    string              `json:"compiler_sha256"`
	LinkFlags         []string            `json:"link_flags"`
	TrainingFlags     []string            `json:"training_flags"`
	TrainingLinkFlags []string            `json:"training_link_flags"`
	Runtime           string              `json:"runtime_sha256"`
	Profile           string              `json:"profile_sha256"`
	Training          string              `json:"training_manifest_sha256"`
	Compiler          string              `json:"compiler"`
	Target            string              `json:"target"`
	Flags             []string            `json:"flags"`
}

func profileHash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }

// RuntimeFingerprint hashes the embedded snapshot, including every header and file boundary.
func RuntimeFingerprint() (string, error) {
	files, err := readRuntime(runtime, "runtime")
	if err != nil {
		return "", err
	}
	return runtimeFingerprint(files), nil
}

func runtimeFingerprint(files []runtimeFile) string {
	h := sha256.New()
	for _, f := range files {
		fmt.Fprintf(h, "%d:%s%d:", len(f.name), f.name, len(f.contents))
		h.Write(f.contents)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// NewStage1ProfileManifest is called by regeneration after llvm-profdata writes the text file.
func NewStage1ProfileManifest(source string, profile []byte, training string) (Stage1ProfileManifest, error) {
	return newStage1ProfileManifest(source, profile, training, Options{Release: true})
}

func newStage1ProfileManifest(source string, profile []byte, training string, options Options) (Stage1ProfileManifest, error) {
	// The text profile has its own content hash. Its local indexed pathname is
	// not a portable build input. Record all other flags from the actual policy.
	options.Profile = ""
	options.ProfileGenerate = false
	snapshot, err := RuntimeFingerprint()
	if err != nil {
		return Stage1ProfileManifest{}, err
	}
	compiler, err := exec.Command("clang", "--version").CombinedOutput()
	if err != nil {
		return Stage1ProfileManifest{}, fmt.Errorf("native: compiler identity: %w", err)
	}
	compilerPath, err := exec.LookPath("clang")
	if err != nil {
		return Stage1ProfileManifest{}, err
	}
	file, err := os.Open(compilerPath)
	if err != nil {
		return Stage1ProfileManifest{}, err
	}
	defer file.Close()
	digest := sha256.New()
	if _, err = io.Copy(digest, file); err != nil {
		return Stage1ProfileManifest{}, err
	}
	trainingOptions := options
	trainingOptions.ProfileGenerate = true
	return Stage1ProfileManifest{
		Version:           2,
		Units:             []Stage1ProfileUnit{{Name: "main.c", SHA256: profileHash([]byte(source))}},
		Runtime:           snapshot,
		Profile:           profileHash(profile),
		Training:          training,
		Compiler:          string(compiler),
		CompilerSHA256:    fmt.Sprintf("%x", digest.Sum(nil)),
		Target:            goruntime.GOOS + "-" + goruntime.GOARCH,
		Flags:             Flags(options),
		LinkFlags:         LinkFlags(options),
		TrainingFlags:     Flags(trainingOptions),
		TrainingLinkFlags: LinkFlags(trainingOptions),
	}, nil
}

func profileMatches(record, current Stage1ProfileManifest) bool {
	return record.Version == current.Version && slices.Equal(record.Units, current.Units) &&
		record.CompilerSHA256 == current.CompilerSHA256 && slices.Equal(record.LinkFlags, current.LinkFlags) &&
		slices.Equal(record.TrainingFlags, current.TrainingFlags) && slices.Equal(record.TrainingLinkFlags, current.TrainingLinkFlags) &&
		record.Runtime == current.Runtime &&
		record.Profile == current.Profile && record.Compiler == current.Compiler && record.Target == current.Target &&
		slices.Equal(record.Flags, current.Flags)
}

func prepareStage1Profile(source string, options Options) (Options, error) {
	if options.Profile == "" {
		return options, nil
	}
	// Even an accidentally supplied profile cannot alter a nonshipping lane.
	if !shippedRelease(options) {
		options.Profile = ""
		return options, nil
	}
	original := options.Profile
	fallback := func(reason string) (Options, error) {
		options.Profile = ""
		fmt.Fprintf(os.Stderr, "adamic: stage 1 profile unavailable (%s); using plain ThinLTO\n", reason)
		return options, nil
	}
	data, err := os.ReadFile(original)
	if err != nil {
		return fallback("missing profile")
	}
	manifest, err := os.ReadFile(filepath.Join(filepath.Dir(original), "manifest.json"))
	if err != nil {
		return fallback("missing manifest")
	}
	var record Stage1ProfileManifest
	if err = json.Unmarshal(manifest, &record); err != nil {
		return fallback("invalid manifest")
	}
	current, err := newStage1ProfileManifest(source, data, record.Training, options)
	if err != nil {
		return options, err
	}
	if !profileMatches(record, current) {
		return fallback("source, runtime, profile or toolchain changed")
	}
	// Hash-addressed immutable indexed copies make runtime.a's existing flag key sound.
	profdata, err := profileDataTool()
	if err != nil {
		return fallback("llvm-profdata missing")
	}
	version, err := exec.Command(profdata, "--version").CombinedOutput()
	if err != nil {
		return fallback("llvm-profdata unavailable")
	}
	cache, err := os.UserCacheDir()
	if err != nil {
		return options, err
	}
	directory := filepath.Join(cache, "adamic", "profiles", profileHash(append(append([]byte{}, data...), version...)))
	if err = os.MkdirAll(directory, 0o755); err != nil {
		return options, err
	}
	indexed := filepath.Join(directory, "profile.profdata")
	if _, err = os.Stat(indexed); err != nil {
		temporary, err := os.MkdirTemp(directory, ".convert-")
		if err != nil {
			return options, err
		}
		defer os.RemoveAll(temporary)
		// Convert the already verified snapshot, never reread a mutable caller path.
		text := filepath.Join(temporary, "profile.txt")
		if err = os.WriteFile(text, data, 0o644); err != nil {
			return options, err
		}
		output := filepath.Join(temporary, "profile.profdata")
		if log, err := exec.Command(profdata, "merge", text, "-o", output).CombinedOutput(); err != nil {
			_ = log
			return fallback("text profile unsupported by llvm-profdata")
		}
		if err = os.Rename(output, indexed); err != nil {
			return options, err
		}
	}
	options.Profile = indexed
	options.profileValidated = true
	return options, nil
}

func profileDataTool() (string, error) {
	compiler, err := exec.LookPath("clang")
	if err != nil {
		return "", err
	}
	if resolved, err := filepath.EvalSymlinks(compiler); err == nil {
		compiler = resolved
	}
	candidate := filepath.Join(filepath.Dir(compiler), "llvm-profdata")
	if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() {
		return candidate, nil
	}
	if found, err := exec.LookPath("llvm-profdata"); err == nil {
		return found, nil
	}
	if goruntime.GOOS == "darwin" {
		if output, err := exec.Command("xcrun", "--find", "llvm-profdata").Output(); err == nil {
			return strings.TrimSpace(string(output)), nil
		}
	}
	return "", fmt.Errorf("llvm-profdata is unavailable")
}
