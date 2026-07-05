package a

// Tags exercises every branch of the jsontag analyzer.
type Tags struct {
	UserID  int    `json:"user_id"`    // snake json — OK
	Name    string `json:"userName"`   // want `is not snake_case`
	Ignored int    `json:"-"`          // dash — skipped
	Omit    int    `json:",omitempty"` // empty name — skipped
	Yaml    string `yaml:"fooBar"`     // want `is not snake_case`
	Plain   string // no tag — skipped
	Other   int    `xml:"x"`           // no json/yaml key — skipped
	Kebab   string `json:"user-id"`    // want `is not snake_case`
	Dotted  string `json:"user.id"`    // want `is not snake_case`
	Spaced  string `json:"user id"`    // want `is not snake_case`
	Leading string `json:"_user_id"`   // want `is not snake_case`
	Doubled string `json:"user__id"`   // want `is not snake_case`
	Accent  string `json:"usér_id"`    // want `is not snake_case`
	Digits  string `json:"sha256_sum"` // snake with digits — OK

	// @-prefixed keys are JSON-LD keywords (W3C Verifiable Credentials etc.)
	// — externally mandated wire syntax, like the `$schema` marker. The "@"
	// prefix itself marks external ownership, so ANY @-prefixed key is
	// sanctioned: external specs own that namespace entirely, and the
	// remainder is never snake_case-checked.
	Context string `json:"@context"`     // JSON-LD keyword — OK
	Type    string `json:"@type"`        // JSON-LD keyword — OK
	Weird   string `json:"@Weird-Thing"` // @-prefixed, non-snake remainder — still OK: the @ namespace is external
	YamlAt  string `yaml:"@context"`     // @-prefix sanction applies to yaml keys too — OK
}
