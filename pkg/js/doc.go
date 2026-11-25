package jsutils

// DocumentedParam Describes one callable exposed to addon JS.
type DocumentedParam struct {
	Name        string
	Type        string // "string", "number", "object", "any", "string[]", ...
	Optional    bool
	Description string
}

type DocumentedFunction struct {
	Name        string // qualified name, e.g. "ek.http.get"
	Self        string // "" for top-level, "ek.http" makes it a namespace member
	Description string // becomes the JSDoc header
	Params      []DocumentedParam
	Returns     string
	ReturnsDoc  string
}
