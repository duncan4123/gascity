package beads

import "testing"

func TestBackendIssueFromBeadIncludesCreateTimeDependencies(t *testing.T) {
	got := backendIssueFromBead(Bead{
		ID:       "gc-child",
		Title:    "Child",
		ParentID: "gc-parent",
		Dependencies: []Dep{{
			DependsOnID: "gc-blocker",
			Type:        "blocks",
		}},
		Needs: []string{"validates:gc-check", "gc-default-blocker"},
	})

	want := []Dep{
		{IssueID: "gc-child", DependsOnID: "gc-blocker", Type: "blocks"},
		{IssueID: "gc-child", DependsOnID: "gc-parent", Type: "parent-child"},
		{IssueID: "gc-child", DependsOnID: "gc-check", Type: "validates"},
		{IssueID: "gc-child", DependsOnID: "gc-default-blocker", Type: "blocks"},
	}
	if len(got.Deps) != len(want) {
		t.Fatalf("deps len = %d, want %d: %#v", len(got.Deps), len(want), got.Deps)
	}
	for i := range want {
		if got.Deps[i] != want[i] {
			t.Fatalf("dep[%d] = %#v, want %#v", i, got.Deps[i], want[i])
		}
	}
}
