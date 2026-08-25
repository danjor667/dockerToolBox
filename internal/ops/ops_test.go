package ops_test

import (
	"context"
	"errors"
	"testing"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/enginetest"
	"dockerToolBox/internal/ops"
)

func running(id, name string) engine.Container {
	return engine.Container{ID: id, Name: name, State: "running"}
}
func stopped(id, name string) engine.Container {
	return engine.Container{ID: id, Name: name, State: "exited"}
}

func TestContainerRemoveAllYesRemovesAll(t *testing.T) {
	f := &enginetest.Fake{Containers: []engine.Container{running("a", "web"), stopped("b", "db")}}
	if err := ops.ContainerRemoveAll(context.Background(), f, engine.Selector{}, ops.Flags{Yes: true}); err != nil {
		t.Fatal(err)
	}
	if len(f.Removed) != 2 {
		t.Fatalf("want 2 removed, got %v", f.Removed)
	}
}

func TestContainerRemoveAllDryRunRemovesNothing(t *testing.T) {
	f := &enginetest.Fake{Containers: []engine.Container{running("a", "web")}}
	if err := ops.ContainerRemoveAll(context.Background(), f, engine.Selector{}, ops.Flags{DryRun: true}); err != nil {
		t.Fatal(err)
	}
	if len(f.Removed) != 0 {
		t.Fatalf("dry-run should remove nothing, got %v", f.Removed)
	}
}

func TestContainerRemoveStoppedOnlyTargetsStopped(t *testing.T) {
	f := &enginetest.Fake{Containers: []engine.Container{running("a", "web"), stopped("b", "db")}}
	if err := ops.ContainerRemoveStopped(context.Background(), f, engine.Selector{}, ops.Flags{Yes: true}); err != nil {
		t.Fatal(err)
	}
	if len(f.Removed) != 1 || f.Removed[0] != "b" {
		t.Fatalf("want only stopped 'b' removed, got %v", f.Removed)
	}
}

func TestContainerRemoveAllHonorsSelector(t *testing.T) {
	f := &enginetest.Fake{Containers: []engine.Container{running("a", "web-1"), running("b", "db-1")}}
	sel := engine.Selector{Names: []string{"web"}}
	if err := ops.ContainerRemoveAll(context.Background(), f, sel, ops.Flags{Yes: true}); err != nil {
		t.Fatal(err)
	}
	if len(f.Removed) != 1 || f.Removed[0] != "a" {
		t.Fatalf("selector should remove only 'a', got %v", f.Removed)
	}
}

// A failure on one container must not abort the whole batch.
func TestContainerRemoveAllContinuesOnError(t *testing.T) {
	f := &enginetest.Fake{
		Containers: []engine.Container{running("a", "web"), running("b", "db")},
		RemoveErr:  map[string]error{"a": errors.New("boom")},
	}
	if err := ops.ContainerRemoveAll(context.Background(), f, engine.Selector{}, ops.Flags{Yes: true}); err != nil {
		t.Fatal(err)
	}
	if len(f.Removed) != 1 || f.Removed[0] != "b" {
		t.Fatalf("want 'b' still removed after 'a' failed, got %v", f.Removed)
	}
}

func TestSystemPruneDryRunTouchesNothing(t *testing.T) {
	f := &enginetest.Fake{}
	if err := ops.SystemPrune(context.Background(), f, ops.Flags{DryRun: true}, true); err != nil {
		t.Fatal(err)
	}
	if f.PrunedCont || f.PrunedImg || f.PrunedNet || f.PrunedVol {
		t.Fatalf("dry-run must not prune anything: %+v", f)
	}
}

func TestSystemPruneWithVolumes(t *testing.T) {
	f := &enginetest.Fake{}
	if err := ops.SystemPrune(context.Background(), f, ops.Flags{Yes: true}, true); err != nil {
		t.Fatal(err)
	}
	if !(f.PrunedCont && f.PrunedImg && f.PrunedNet && f.PrunedVol) {
		t.Fatalf("prune --volumes should touch all four: %+v", f)
	}
}

func TestSystemPruneWithoutVolumesSkipsVolumes(t *testing.T) {
	f := &enginetest.Fake{}
	if err := ops.SystemPrune(context.Background(), f, ops.Flags{Yes: true}, false); err != nil {
		t.Fatal(err)
	}
	if f.PrunedVol {
		t.Fatal("prune without --volumes must not prune volumes")
	}
}
