package command

// Param describes one command parameter for form generation and validation.
type Param struct {
	Name     string
	Required bool
	Variadic bool
}

// Spec is the normalized metadata used by discovery, search and command forms.
type Spec struct {
	Name     string
	Arity    int
	Flags    []string
	Summary  string
	Category string
	Params   []Param
}
