package main

import "testing"

func TestBuildBreadcrumb(t *testing.T) {
	root := testRoot(t)

	cases := []struct {
		relPath string
		want    []breadcrumbItem
	}{
		{
			relPath: "index.md",
			want:    []breadcrumbItem{{Name: "index", Href: "/"}},
		},
		{
			relPath: "log.md",
			want: []breadcrumbItem{
				{Name: "index", Href: "/"},
				{Name: "log"},
			},
		},
		{
			relPath: "widgets/index.md",
			want: []breadcrumbItem{
				{Name: "index", Href: "/"},
				{Name: "widgets", Href: "/view/widgets/index.md"},
				{Name: "index"},
			},
		},
		{
			relPath: "widgets/gadget.md",
			want: []breadcrumbItem{
				{Name: "index", Href: "/"},
				{Name: "widgets", Href: "/view/widgets/index.md"},
				{Name: "gadget"},
			},
		},
	}

	for _, c := range cases {
		got := buildBreadcrumb(root, c.relPath)
		if len(got) != len(c.want) {
			t.Fatalf("buildBreadcrumb(%q) = %+v, want %+v", c.relPath, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("buildBreadcrumb(%q)[%d] = %+v, want %+v", c.relPath, i, got[i], c.want[i])
			}
		}
	}
}
