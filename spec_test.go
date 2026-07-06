package spec

import (
	"testing"

	"github.com/alecthomas/kong"
)

type testCLI struct {
	BoolWithDefault    bool `default:"true"`
	BoolWithoutDefault bool
	BoolZeroDefault    bool   `default:"false"`
	IntWithDefault     int    `default:"42"`
	IntZeroDefault     int    `default:"0"`
	StringWithDefault  string `default:"hello"`
	StringZeroDefault  string `default:""`
}

func TestZeroDefaultSuppressed(t *testing.T) {
	k, err := kong.New(&testCLI{}, kong.NoDefaultHelp())
	if err != nil {
		t.Fatalf("kong.New: %v", err)
	}

	cmd := Command(k.Model.Node)

	tests := []struct {
		flagName      string
		shouldHaveDef bool
	}{
		{"bool-with-default", true},     // default:"true" != zero (false)
		{"bool-without-default", false}, // no default tag
		{"bool-zero-default", false},    // default:"false" == zero (false)
		{"int-with-default", true},      // default:"42" != zero (0)
		{"int-zero-default", false},     // default:"0" == zero (0)
		{"string-with-default", true},   // default:"hello" != zero ("")
		{"string-zero-default", false},  // default:"" == zero ("")
	}

	for _, tt := range tests {
		t.Run(tt.flagName, func(t *testing.T) {
			key := "--" + tt.flagName
			flag, ok := cmd.Flags[key]
			if !ok {
				key = "--" + tt.flagName + "="
				flag, ok = cmd.Flags[key]
			}
			if !ok {
				t.Fatalf("flag --%s not found in cmd.Flags", tt.flagName)
			}
			hasDef := flag.Default != ""
			if hasDef != tt.shouldHaveDef {
				t.Errorf("flag --%s: expected Default=%v, got Default=%q",
					tt.flagName, tt.shouldHaveDef, flag.Default)
			}
		})
	}
}
