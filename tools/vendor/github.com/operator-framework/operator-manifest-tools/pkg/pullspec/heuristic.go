package pullspec

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"
	"unicode"
)

const (
	// Alphanumeric characters
	alnum = "[a-zA-Z0-9]"
	// Characters that you might find in a registry, namespace, repo or tag
	name = `[a-zA-Z0-9\-._]`
	// Base16 characters
	base16 = "[a-fA-F0-9]"
)

// rules
// nolint:unused,deadcode,varcheck
var (
	templates = template.Must(template.New("all").Parse(""))
	// A basic name is anything that contains only alphanumeric and name
	// characters and starts and ends with an alphanumeric character
	basicName = mustCompileRule(templates, "basicName",
		"(?:(?:{{ .alnum }}{{ .name }}*{{ .alnum }})|{{ .alnum }})")

	// A named tag is ':' followed by a basic name
	namedTag = mustCompileRule(templates, "namedTag", `(?::{{ template "basicName" . }})`)
	// A digest is "@sha256:" followed by exactly 64 base16 characters
	digest = mustCompileRule(templates, "digest", `(?:@sha256:{{ .base16 }}{{ print "{64}"}})`)

	// A tag is either a named tag or a digest
	tag = mustCompileRule(templates, "tag", `(?:{{ template "namedTag" . }}|{{ template "digest" . }})`)

	// Registry is a basic name that contains at least one dot
	// followed by an optional port number
	registry = mustCompileRule(templates, "registry",
		`(?:{{ .alnum }}{{ .name }}*\.{{ .name }}*{{ .alnum }}(?::\d+)?)`)

	// Namespace is a basic name
	namespace = mustCompileRule(templates, "namespace", `{{ template "basicName" . }}`)

	// Repo is a basic name followed by a tag
	// NOTE: Tag is REQUIRED, otherwise regex picks up too many false positives,
	// such as URLs, math and possibly many others.
	repo = mustCompileRule(templates, "repo", `{{ template "basicName" . }}{{ template "tag" . }}`)

	// Pullspec is registry/namespace*/repo
	pullspecRule = mustCompileRule(templates, "pullspec", `{{ template "registry" . }}/(?:{{ template "namespace" . }}/)*{{ template "repo" . }}`)
)

// regexes
// nolint:unused,deadcode
var (
	pullspec  = regexp.MustCompile(mustExecute(templates, "{{template `pullspec` .}}", "alnum", alnum, "name", name, "base16", base16))
	candidate = regexp.MustCompile(`[a-zA-Z0-9/\-\._@:]+`)
	full      = regexp.MustCompile(mustExecute(templates, `^{{ template "pullspec" . }}$`, "alnum", alnum, "name", name, "base16", base16))
)

// mustExecute executes a template, panicing on error
func mustExecute(templates *template.Template, templ string, data ...interface{}) string {
	if len(data)%2 != 0 {
		panic("data length must be 2")
	}

	templateData := map[string]interface{}{}

	for i := 0; i < len(data); i = i + 2 {
		key := data[i].(string)
		templateData[key] = data[i+1]
	}

	strb := strings.Builder{}

	err := template.Must(templates.Parse(templ)).Execute(&strb, templateData)

	if err != nil {
		panic(err)
	}

	return strb.String()
}

// mustCompileRule compiles a template rule and panics on error
func mustCompileRule(templates *template.Template, name string, templ string) *template.Template {
	return template.Must(
		templates.Parse(
			fmt.Sprintf(`{{ define "%s" }}%s{{ end }}`, name, templ),
		))
}

// Heuristic takes a text and returns a matching slice of
// []{start, end} slices of substring indices
type Heuristic func(text string) [][]int

// DefaultHeuristic attempts to find all pullspecs in arbitrary structured/unstructured text.
// Returns a list of (start, end) tuples such that:
//     text[start:end] == <n-th pullspec in text> for all (start, end)
// The basic idea:
// - Find continuous sequences of characters that might appear in a pullspec
//   - That being <alphanumeric> + "/-._@:"
// - For each such sequence:
//   - Strip non-alphanumeric characters from both ends
//   - Match remainder against the pullspec regex
// Put simply, this heuristic should find anything in the form:
//     registry/namespace*/repo:tag
//     registry/namespace*/repo@sha256:digest
// Where registry must contain at least one '.' and all parts follow various
// restrictions on the format (most typical pullspecs should be caught). Any
// number of namespaces, including 0, is valid.
// NOTE: Pullspecs without a tag (implicitly :latest) will not be caught.
// This would produce way too many false positives (and 1 false positive
// is already too many).
// :param text: Arbitrary blob of text in which to find pullspecs
// :return: Slice of []int{start, end} tuples of substring indices
func DefaultHeuristic(text string) [][]int {
	pullspecs := [][]int{}

	candidates := pullspecCandidates(text)
	for _, bounds := range candidates {
		i, j := bounds[0], bounds[1]
		i, j = adjustForArbitraryText(text, i, j)

		candidate := text[i:j]
		if len(candidate) != 0 && full.Match([]byte(candidate)) {
			pullspecs = append(pullspecs, []int{i, j})
			log.Printf("Pull spec heuristic: %s looks like a pullspec\n", candidate)
		}
	}

	return pullspecs
}

func pullspecCandidates(text string) [][]int {
	return candidate.FindAllIndex([]byte(text), -1)
}

func adjustForArbitraryText(text string, i, j int) (int, int) {
	// Strip all non-alphanumeric characters from start and end of pullspec
	// candidate to account for various structured/unstructured text elements
	for i < len(text) && !isAlphaNum(rune(text[i])) && i != j {
		i = i + 1
	}
	for j > 0 && !isAlphaNum(rune(text[j-1])) && j != i {
		j = j - 1
	}

	return i, j
}

func isAlphaNum(s rune) bool {
	return unicode.IsLetter(s) || unicode.IsDigit(s)
}
