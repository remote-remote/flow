package notes

import "testing"

func TestInsertUnderHeading(t *testing.T) {
	tests := []struct {
		name    string
		content string
		heading string
		want    string
	}{
		{
			name:    "inserts directly below the heading",
			content: "# Day\n\n## Tasks\n- old\n\n## Notes\n",
			heading: "## Tasks",
			want:    "# Day\n\n## Tasks\n- new\n- old\n\n## Notes\n",
		},
		{
			name:    "appends the section when the heading is absent",
			content: "# Day\n",
			heading: "## Tasks",
			want:    "# Day\n\n## Tasks\n- new\n",
		},
		{
			name:    "ignores a deeper heading with the same name",
			content: "# Day\n\n### Tasks\n\n## Tasks\n",
			heading: "## Tasks",
			want:    "# Day\n\n### Tasks\n\n## Tasks\n- new\n",
		},
		{
			name:    "ignores the heading text inside prose",
			content: "# Day\n\nwrote up ## Tasks today\n\n## Tasks\n",
			heading: "## Tasks",
			want:    "# Day\n\nwrote up ## Tasks today\n\n## Tasks\n- new\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := insertUnderHeading(tt.content, tt.heading, "- new"); got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := map[string]string{
		"How does auth work?":    "how-does-auth-work",
		"  Agentic Flow  ":       "agentic-flow",
		"flow/publish -> Linear": "flow-publish-linear",
	}
	for in, want := range tests {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q) = %q, want %q", in, got, want)
		}
	}
}
