package handler

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestCrud_NoArgsDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Crud panicked with no args: %v", r)
		}
	}()
	Crud(&cobra.Command{}, nil, NewFlags(nil))
}
