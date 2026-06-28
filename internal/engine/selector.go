package engine

import "strings"

// Selector filters Docker resources by name and/or label. It is applied
// client-side so the same semantics work uniformly across containers,
// images, volumes, and networks.
//
//   - Names: substring match, OR semantics — a resource matches if any of
//     its names/tags contains any of the patterns. Empty means "any".
//   - Labels: AND semantics — a resource matches only if it carries every
//     listed label. A map value of "" matches on key presence alone;
//     otherwise the value must be equal.
type Selector struct {
	Names  []string
	Labels map[string]string
}

// Empty reports whether the selector would match everything.
func (s Selector) Empty() bool {
	return len(s.Names) == 0 && len(s.Labels) == 0
}

func (s Selector) matchName(names ...string) bool {
	if len(s.Names) == 0 {
		return true
	}
	for _, pat := range s.Names {
		for _, n := range names {
			if n != "" && strings.Contains(n, pat) {
				return true
			}
		}
	}
	return false
}

func (s Selector) matchLabels(labels map[string]string) bool {
	for k, want := range s.Labels {
		got, ok := labels[k]
		if !ok {
			return false
		}
		if want != "" && got != want {
			return false
		}
	}
	return true
}

// FilterContainers returns the containers that match the selector.
func (s Selector) FilterContainers(in []Container) []Container {
	if s.Empty() {
		return in
	}
	var out []Container
	for _, c := range in {
		if s.matchName(c.Name) && s.matchLabels(c.Labels) {
			out = append(out, c)
		}
	}
	return out
}

// FilterImages returns the images that match the selector.
func (s Selector) FilterImages(in []Image) []Image {
	if s.Empty() {
		return in
	}
	var out []Image
	for _, im := range in {
		if s.matchName(im.Tags...) && s.matchLabels(im.Labels) {
			out = append(out, im)
		}
	}
	return out
}

// FilterVolumes returns the volumes that match the selector.
func (s Selector) FilterVolumes(in []Volume) []Volume {
	if s.Empty() {
		return in
	}
	var out []Volume
	for _, v := range in {
		if s.matchName(v.Name) && s.matchLabels(v.Labels) {
			out = append(out, v)
		}
	}
	return out
}

// FilterNetworks returns the networks that match the selector.
func (s Selector) FilterNetworks(in []Network) []Network {
	if s.Empty() {
		return in
	}
	var out []Network
	for _, n := range in {
		if s.matchName(n.Name) && s.matchLabels(n.Labels) {
			out = append(out, n)
		}
	}
	return out
}
