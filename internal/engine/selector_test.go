package engine

import "testing"

func TestSelectorEmptyMatchesAll(t *testing.T) {
	cs := []Container{{Name: "a"}, {Name: "b"}}
	got := Selector{}.FilterContainers(cs)
	if len(got) != 2 {
		t.Fatalf("empty selector should match all; got %d of 2", len(got))
	}
}

func TestSelectorNameSubstringOR(t *testing.T) {
	cs := []Container{
		{Name: "web-1"},
		{Name: "web-2"},
		{Name: "db-1"},
	}
	got := Selector{Names: []string{"web"}}.FilterContainers(cs)
	if len(got) != 2 {
		t.Fatalf("name=web should match 2; got %d", len(got))
	}

	got = Selector{Names: []string{"web", "db"}}.FilterContainers(cs)
	if len(got) != 3 {
		t.Fatalf("name=web,db (OR) should match 3; got %d", len(got))
	}
}

func TestSelectorLabelAND(t *testing.T) {
	cs := []Container{
		{Name: "a", Labels: map[string]string{"env": "dev", "tier": "web"}},
		{Name: "b", Labels: map[string]string{"env": "dev"}},
		{Name: "c", Labels: map[string]string{"env": "prod", "tier": "web"}},
	}

	// key presence only
	if got := (Selector{Labels: map[string]string{"tier": ""}}).FilterContainers(cs); len(got) != 2 {
		t.Fatalf("label tier (presence) should match 2; got %d", len(got))
	}
	// key=value
	if got := (Selector{Labels: map[string]string{"env": "dev"}}).FilterContainers(cs); len(got) != 2 {
		t.Fatalf("label env=dev should match 2; got %d", len(got))
	}
	// AND across two labels
	sel := Selector{Labels: map[string]string{"env": "dev", "tier": "web"}}
	if got := sel.FilterContainers(cs); len(got) != 1 || got[0].Name != "a" {
		t.Fatalf("env=dev AND tier=web should match only 'a'; got %v", got)
	}
}

func TestSelectorNameAndLabelCombined(t *testing.T) {
	cs := []Container{
		{Name: "web-dev", Labels: map[string]string{"env": "dev"}},
		{Name: "web-prod", Labels: map[string]string{"env": "prod"}},
	}
	sel := Selector{Names: []string{"web"}, Labels: map[string]string{"env": "dev"}}
	got := sel.FilterContainers(cs)
	if len(got) != 1 || got[0].Name != "web-dev" {
		t.Fatalf("name=web AND env=dev should match only 'web-dev'; got %v", got)
	}
}

func TestSelectorImageMatchesByTag(t *testing.T) {
	ims := []Image{
		{Tags: []string{"nginx:latest"}},
		{Tags: []string{"redis:7"}},
		{Tags: nil}, // untagged
	}
	got := Selector{Names: []string{"nginx"}}.FilterImages(ims)
	if len(got) != 1 {
		t.Fatalf("name=nginx should match 1 image; got %d", len(got))
	}
}
