package ant

import (
	"reflect"
	"testing"
)

func TestFindMentions(t *testing.T) {
	cases := []struct {
		name   string
		prefix string
		body   string
		want   []string
	}{
		{"plain ids in order, deduped", "demo",
			"see demo-AbCdE and demo-XyZ23, then demo-AbCdE again", []string{"demo-AbCdE", "demo-XyZ23"}},
		{"punctuation around ids", "demo",
			"(`demo-AbCdE`), \"demo-XyZ23\"; end with demo-Qwert.", []string{"demo-AbCdE", "demo-XyZ23", "demo-Qwert"}},
		{"other prefixes ignored", "demo",
			"ait-AbCdE and other-AbCdE", nil},
		{"dotted ait child ids ignored", "demo",
			"epic demo-AbCdE.3.1 and child demo-XyZ23.2", nil},
		{"shorter than the minimum length ignored", "demo",
			"demo-abcd", nil},
		{"part of a longer token ignored", "demo",
			"xdemo-AbCdE demo-AbCdE_x demo-AbCdE-y", nil},
		{"characters outside the alphabet end no match", "demo",
			"demo-AbCdE0", nil},
		{"prefix with regex metacharacters", "my.app",
			"my.app-AbCdE but not myxapp-AbCdE", []string{"my.app-AbCdE"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := findMentions(c.prefix, c.body)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("findMentions(%q, %q) = %q, want %q", c.prefix, c.body, got, c.want)
			}
		})
	}
}
