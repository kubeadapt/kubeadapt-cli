package update

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{"newer minor", "0.2.0", "0.1.4", true},
		{"newer patch", "0.1.5", "0.1.4", true},
		{"newer major", "1.0.0", "0.99.99", true},
		{"large numbers", "10.0.0", "9.999.999", true},
		{"equal", "0.2.0", "0.2.0", false},
		{"older minor", "0.1.4", "0.2.0", false},
		{"older patch", "0.1.3", "0.1.4", false},
		{"v-prefix on both", "v0.2.0", "v0.1.4", true},
		{"v-prefix on latest only", "v0.2.0", "0.1.4", true},
		{"v-prefix on current only", "0.2.0", "v0.1.4", true},
		{"prerelease ignored on latest", "0.2.0-rc1", "0.1.4", true},
		{"prerelease ignored on current", "0.2.0", "0.2.0-rc1", false},
		{"build metadata ignored", "0.2.0+abc", "0.1.4", true},
		{"empty latest", "", "0.1.4", false},
		{"empty current", "0.2.0", "", false},
		{"malformed latest", "abc", "0.1.4", false},
		{"malformed current", "0.2.0", "abc.def", false},
		{"two-component version", "0.2", "0.1.4", false},
		{"four-component version", "0.2.0.1", "0.1.4", false},
		{"negative number", "-1.0.0", "0.1.4", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := isNewer(tc.latest, tc.current)
			if got != tc.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
			}
		})
	}
}
