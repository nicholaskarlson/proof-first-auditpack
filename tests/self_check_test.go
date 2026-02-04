package tests

import (
	"testing"

	"github.com/nicholaskarlson/proof-first-auditpack/internal/auditpack"
)

func TestSelfCheck_OK(t *testing.T) {
	t.Parallel()

	_, err := auditpack.SelfCheck(auditpack.SelfCheckOptions{
		Strict: true,
		Keep:   false,
	})
	if err != nil {
		t.Fatalf("self-check failed: %v", err)
	}
}
