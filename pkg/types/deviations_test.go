package types

import "testing"

func TestIntentDeviations_IntentName(t *testing.T) {
	cases := map[string]struct {
		typ  DeviationType
		name string
		want string
	}{
		"config prefixed": {DeviationTypeConfig, "config-intent1-srl-srl1", "intent1-srl-srl1"},
		"target prefixed": {DeviationTypeTarget, "target-srl1", "srl1"},
		// legacy CRs predate the prefix; return the raw name
		"no known prefix": {DeviationTypeConfig, "intent1-srl-srl1", "intent1-srl-srl1"},
		// only the leading prefix is stripped, even when the resource itself begins with one
		"resource name starts with prefix literal": {DeviationTypeConfig, "config-config-name", "config-name"},
		// don't strip a prefix that doesn't match this IntentDeviations' type
		"type mismatch returns raw": {DeviationTypeUnknown, "config-intent1", "config-intent1"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := NewDeviations("target-x", tc.name, tc.typ, 0)
			if got := d.IntentName(); got != tc.want {
				t.Fatalf("IntentName() = %q, want %q", got, tc.want)
			}
		})
	}
}
