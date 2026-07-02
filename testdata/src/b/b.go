// Package b mirrors an external self-describing JSON-Schema document: the
// $schema field marks it, so its camelCase keys are exempt.
package b

type envelope struct {
	Schema  string `json:"$schema"`
	Version string `json:"version"`
	Runs    []run  `json:"runs"`
}

type run struct {
	ShortDescription string `json:"shortDescription"`
	HelpURI          string `json:"helpUri,omitempty"`
}

func use() { _ = envelope{}; _ = run{} }
