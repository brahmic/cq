package version

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "devel", input: "(devel)", want: ""},
		{name: "tagged", input: "v1.2.3", want: "1.2.3"},
		{name: "plain", input: "1.2.3", want: "1.2.3"},
		{name: "pseudo", input: "v0.0.0-20260215193510-f10f24864b3b+dirty", want: "0.0.0-20260215193510-f10f24864b3b+dirty"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalize(tc.input); got != tc.want {
				t.Fatalf("normalize(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}
