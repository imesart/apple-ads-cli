package structure

import "testing"

func TestDeriveAppNameShortValue_UsesRedactionSeparators(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "colon", in: "FitTrack SE: fitness calories", want: "FitTrack SE"},
		{name: "slash", in: "FitTrack / Search", want: "FitTrack"},
		{name: "paren", in: "FitTrack (Search)", want: "FitTrack"},
		{name: "dash", in: "FitTrack - Search", want: "FitTrack"},
		{name: "bullet not a separator", in: "FitTrack • Search", want: "FitTrack • Search"},
		{name: "plain spaces are not separators", in: "FitTrack Search", want: "FitTrack Search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deriveAppNameShortValue(tt.in); got != tt.want {
				t.Fatalf("deriveAppNameShortValue(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
