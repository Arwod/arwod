package security_test

import (
	"fmt"
	"regexp"
	"regexp/syntax"
	"testing"

	"github.com/pocketbase/pocketbase/tools/security"
)

func TestRandomStringByRegex(t *testing.T) {
	scenarios := []struct {
		pattern     string
		flags       []syntax.Flags
		expectError bool
	}{
		{``, nil, true},
		{`test`, nil, false},
		{`\d+`, []syntax.Flags{syntax.POSIX}, true},
		{`\d+`, nil, false},
		{`\d*`, nil, false},
		{`\d{1,10}`, nil, false},
		{`\d{3}`, nil, false},
		{`\d{0,}-abc`, nil, false},
		{`[a-zA-Z_]*`, nil, false},
		{`[^a-zA-Z]{5,30}`, nil, false},
		{`\w+_abc`, nil, false},
		{pattern: `[2-9]{5}-\w+`, flags: nil, expectError: false},
		{`(a|b|c)`, nil, false},
	}

	for i, s := range scenarios {
		t.Run(fmt.Sprintf("%d_%q", i, s.pattern), func(t *testing.T) {
			// Test multiple generations to ensure consistency
			for j := 0; j < 5; j++ {
				str, err := security.RandomStringByRegex(s.pattern, s.flags...)

				hasErr := err != nil
				if hasErr != s.expectError {
					t.Fatalf("Expected hasErr %v, got %v (%v)", s.expectError, hasErr, err)
				}

				if hasErr {
					return
				}

				r, err := regexp.Compile(s.pattern)
				if err != nil {
					t.Fatal(err)
				}

				if !r.Match([]byte(str)) {
					t.Fatalf("Expected %q to match pattern %v", str, s.pattern)
				}
			}
		})
	}
}
