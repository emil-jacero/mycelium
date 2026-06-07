package dag

import (
	"errors"
	"reflect"
	"testing"
)

func TestTopoSortOrders(t *testing.T) {
	g := New[string]()
	for _, id := range []string{"fetch", "summarize", "report"} {
		if err := g.AddNode(id, id); err != nil {
			t.Fatal(err)
		}
	}
	// report -> summarize -> fetch (report depends on summarize, etc.)
	must(t, g.DependsOn("summarize", "fetch"))
	must(t, g.DependsOn("report", "summarize"))

	order, err := g.TopoSort()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"fetch", "summarize", "report"}
	if !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
}

func TestTopoSortDeterministic(t *testing.T) {
	g := New[string]()
	// Three independent roots must come out alphabetically, every time.
	for _, id := range []string{"c", "a", "b"} {
		must(t, g.AddNode(id, id))
	}
	order, err := g.TopoSort()
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"a", "b", "c"}; !reflect.DeepEqual(order, want) {
		t.Fatalf("order = %v, want %v", order, want)
	}
}

func TestCycleDetected(t *testing.T) {
	g := New[string]()
	must(t, g.AddNode("a", "a"))
	must(t, g.AddNode("b", "b"))
	must(t, g.DependsOn("a", "b"))
	must(t, g.DependsOn("b", "a"))

	if _, err := g.TopoSort(); !errors.Is(err, ErrCycle) {
		t.Fatalf("expected ErrCycle, got %v", err)
	}
}

func TestUnknownDependency(t *testing.T) {
	g := New[string]()
	must(t, g.AddNode("a", "a"))
	if err := g.DependsOn("a", "ghost"); err == nil {
		t.Fatal("expected error for unknown dependency")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
