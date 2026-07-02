package jsontag_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/tools/go/analysis/analysistest"

	jsontag "github.com/gomatic/yze-go-jsontag"
)

func TestNonSnakeCaseTagsAreReported(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), jsontag.Analyzer, "a", "b", "c")
}

func TestRegistrationIsWellFormed(t *testing.T) {
	assert.NoError(t, jsontag.Registration.Validate())
	assert.Equal(t, "yze/jsontag", jsontag.Registration.RuleID())
	assert.Same(t, jsontag.Analyzer, jsontag.Registration.Analyzer)
}
