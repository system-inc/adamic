package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type negativeWitness struct {
	// Form is empty for the literal type lies, or checkedBoundaryForm.
	Form              string `json:"form,omitempty"`
	RulingTask        string `json:"ruling_task"`
	Ruling            string `json:"ruling,omitempty"`
	Declaration       string `json:"declaration"`
	DeclaredType      string `json:"declared_type"`
	RepairFrom        string `json:"repair_from"`
	RepairTo          string `json:"repair_to"`
	TypeCorrectSHA256 string `json:"type_correct_sha256"`
	Exit              int    `json:"exit"`
	Stderr            string `json:"stderr"`
}
type negativeWitnessList struct {
	Comment   string                     `json:"_comment"`
	Witnesses map[string]negativeWitness `json:"witnesses"`
}

func contentHash(source []byte) string {
	hash := sha256.Sum256(source)
	return fmt.Sprintf("%x", hash)
}
func validContentHash(hash string) bool {
	decoded, err := hex.DecodeString(hash)
	return err == nil && len(decoded) == sha256.Size && hash == strings.ToLower(hash)
}
func readNegativeWitnesses(path string) (negativeWitnessList, error) {
	var list negativeWitnessList
	data, err := os.ReadFile(path)
	if err != nil {
		return list, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&list); err != nil {
		return list, err
	}
	var extra any
	if err = decoder.Decode(&extra); err != io.EOF {
		return list, fmt.Errorf("negative witness list has trailing JSON")
	}
	if !strings.Contains(list.Comment, "Adding an entry is a ruling") || !strings.Contains(list.Comment, "@system_adamic") || list.Witnesses == nil {
		return list, fmt.Errorf("negative witness list needs its ruling header and witnesses object")
	}
	for hash, witness := range list.Witnesses {
		if !validContentHash(hash) || !validContentHash(witness.TypeCorrectSHA256) || witness.TypeCorrectSHA256 == hash {
			return list, fmt.Errorf("invalid negative or type-correct content hash %q", hash)
		}
		if _, listed := list.Witnesses[witness.TypeCorrectSHA256]; listed {
			return list, fmt.Errorf("type-correct program cannot be listed: %s", witness.TypeCorrectSHA256)
		}
		if !strings.HasPrefix(witness.RulingTask, "#") || len(witness.RulingTask) < 2 || witness.Exit != 70 || !strings.HasPrefix(witness.Stderr, "adamic: panic: ") || !strings.HasSuffix(witness.Stderr, "\n") || strings.Count(witness.Stderr, "\n") != 1 {
			return list, fmt.Errorf("negative witness needs ruling task, exit 70 and exact stop message: %s", hash)
		}
		if _, err := declarationLine(witness.Declaration); err != nil {
			return list, err
		}
		if !supportedLie(witness) {
			return list, fmt.Errorf("unsupported type lie or repair: %s", hash)
		}
	}
	return list, nil
}

// checkedBoundaryForm is a type lie a checked boundary catches (#995z2fa, Adamic, Oct 25): a value a
// checked cast let in, stopped at a checked view's own read or write with that check's located message.
// Its repair is arbitrary text, so the run holds the repaired copy to Node as well.
const checkedBoundaryForm = "checked_boundary"

// checkedBoundaryStop is the view check's message, and only it: a compiler-bug panic or any other guard
// never qualifies. Its expected and found types are what the declaration is tied to.
var checkedBoundaryStop = regexp.MustCompile(`^adamic: panic: field read failed: [^\n]* expected ([^,\n]+), found ([^\n]+)\n$`)

func supportedLie(witness negativeWitness) bool {
	switch witness.Form {
	case "":
		return supportedLiteralLie(witness)
	case checkedBoundaryForm:
		return supportedCheckedBoundary(witness)
	}
	return false
}

// A checked boundary's entry names its ruling, and the declared type it states is the one the stop
// expected, for a read, or found in the slot, for a write.
func supportedCheckedBoundary(witness negativeWitness) bool {
	stop := checkedBoundaryStop.FindStringSubmatch(witness.Stderr)
	if witness.Ruling == "" || stop == nil || strings.Contains(witness.Stderr, "compiler bug") || witness.DeclaredType == "" {
		return false
	}
	if witness.DeclaredType != stop[1] && witness.DeclaredType != stop[2] {
		return false
	}
	return witness.RepairFrom != "" && witness.RepairTo != "" && witness.RepairFrom != witness.RepairTo
}

// declaresRuledType says the declaration line states the stop's expected type: a declaration of exactly
// that type, or an optional field, whose read expects its declared type plus undefined (#1yr7rdh, Adamic,
// Oct 28). A required field's read expects exactly its declared type.
func declaresRuledType(declaration, declared string) bool {
	if regexp.MustCompile(`: ` + regexp.QuoteMeta(declared) + `[ ;]`).MatchString(declaration) {
		return true
	}
	present, optional := strings.CutSuffix(declared, " | undefined")
	return optional && regexp.MustCompile(`\?: `+regexp.QuoteMeta(present)+`(;|,| })`).MatchString(declaration)
}

// repairedSource is the type-correct copy: every RepairFrom replaced with RepairTo.
func repairedSource(source []byte, witness negativeWitness) []byte {
	return bytes.ReplaceAll(source, []byte(witness.RepairFrom), []byte(witness.RepairTo))
}
func declarationLine(declaration string) (int, error) {
	at := strings.LastIndex(declaration, ":")
	if at <= 0 {
		return 0, fmt.Errorf("declaration must name file:line")
	}
	line, err := strconv.Atoi(declaration[at+1:])
	if err != nil || line < 1 {
		return 0, fmt.Errorf("declaration must name file:line")
	}
	return line, nil
}

// These literal contradictions do not ask the TypeScript checker to prove its own assertions.
// Other forms need a new ruling and a corresponding independent evidence validator.
func supportedLiteralLie(witness negativeWitness) bool {
	switch witness.DeclaredType {
	case "string":
		return (witness.RepairFrom == "undefined!" || witness.RepairFrom == "null!") && witness.RepairTo == `"ready"`
	case "number":
		return witness.RepairFrom == "return 'wrong';" && witness.RepairTo == "return 1;"
	}
	return false
}
func validateWitnessSource(source []byte, witness negativeWitness) error {
	if regexp.MustCompile(`\b(import|export)\b`).Match(source) {
		return fmt.Errorf("negative witnesses must be self-contained")
	}
	if !supportedLie(witness) {
		return fmt.Errorf("unsupported type lie or repair")
	}
	line, err := declarationLine(witness.Declaration)
	if err != nil {
		return err
	}
	lines := strings.Split(string(source), "\n")
	if line > len(lines) {
		return fmt.Errorf("violated declaration is outside the source")
	}
	declaration := lines[line-1]
	if witness.Form == checkedBoundaryForm {
		if !declaresRuledType(declaration, witness.DeclaredType) {
			return fmt.Errorf("violated declaration does not declare its ruled type")
		}
		if bytes.Count(source, []byte(witness.RepairFrom)) != 1 {
			return fmt.Errorf("a checked boundary's repair must replace exactly one place")
		}
	} else if witness.DeclaredType == "string" {
		pattern := `^(let [A-Za-z_$][A-Za-z0-9_$]*: string = ` + regexp.QuoteMeta(witness.RepairFrom) + `;|class [A-Za-z_$][A-Za-z0-9_$]* \{ [A-Za-z_$][A-Za-z0-9_$]*: string = ` + regexp.QuoteMeta(witness.RepairFrom) + `; \})$`
		if !regexp.MustCompile(pattern).MatchString(declaration) {
			return fmt.Errorf("declared string does not contain its ruled literal type lie")
		}
	} else if string(source) != "function lookAhead(callback:()=>unknown):unknown {return callback();}\nfunction misfit():unknown {return 'wrong';}\nconst result=lookAhead(misfit) as number;\nconsole.log(String(result));\n" || line != 3 {
		return fmt.Errorf("declared number does not contain its ruled literal callback type lie")
	}
	repaired := repairedSource(source, witness)
	if bytes.Equal(source, repaired) || contentHash(repaired) != witness.TypeCorrectSHA256 {
		return fmt.Errorf("type-correct repair hash differs")
	}
	return nil
}
func acceptsNegativeWitness(witness negativeWitness, root string, node, javascript, native observation) bool {
	// Source Node must finish. Engine failures belong to ordinary Node agreement.
	if node.Error != "" || node.Exit != 0 {
		return false
	}
	message := strings.ReplaceAll(witness.Stderr, "{root}", root)
	for _, backend := range []observation{javascript, native} {
		if backend.Error != "" || backend.Exit != witness.Exit || backend.Stderr != message {
			return false
		}
	}
	// A listed stop is still required to agree between the backends on stdout.
	return javascript.Stdout == native.Stdout
}

// Paths schedule rechecks only; exemption still comes exclusively from contentHash.
// Otherwise a repaired placeholder already admitted on base could evade the oracle.
func negativeRelated(path string, source []byte, list negativeWitnessList) bool {
	hash := contentHash(source)
	for ruledHash, witness := range list.Witnesses {
		at := strings.LastIndex(witness.Declaration, ":")
		if hash == ruledHash || hash == witness.TypeCorrectSHA256 || (at > 0 && path == witness.Declaration[:at]) {
			return true
		}
	}
	return false
}
