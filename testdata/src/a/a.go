package a

// Tags exercises every branch of the jsontag analyzer.
type Tags struct {
	UserID  int    `json:"user_id"`    // snake json — OK
	Name    string `json:"userName"`   // want `is not snake_case`
	Ignored int    `json:"-"`          // dash — skipped
	Omit    int    `json:",omitempty"` // empty name — skipped
	Yaml    string `yaml:"fooBar"`     // want `is not snake_case`
	Plain   string // no tag — skipped
	Other   int    `xml:"x"` // no json/yaml key — skipped
}
