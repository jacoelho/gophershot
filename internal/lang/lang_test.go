package lang

import "testing"

func TestFromInputPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		path string
		want Language
	}{
		{path: "", want: LanguageGo},
		{path: "-", want: LanguageGo},
		{path: "main.go", want: LanguageGo},
		{path: "infra.tf", want: LanguageTerraform},
		{path: "infra.TF", want: LanguageTerraform},
	}

	for _, tc := range cases {
		if got := FromInputPath(tc.path); got != tc.want {
			t.Fatalf("FromInputPath(%q) = %q, want %q", tc.path, got, tc.want)
		}
	}
}
