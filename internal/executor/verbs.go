package executor

import (
	"context"

	"dockerToolBox/internal/engine"
	"dockerToolBox/internal/ops"
)

// verb runs a built-in operation through the engine, honoring the safety
// flags. Verbs are the SDK-native form of a custom-command step: they run
// with no shell and no --allow-shell.
type verb func(ctx context.Context, eng engine.API, f ops.Flags) error

// verbs maps a scalar step string to its native operation. The keys mirror
// the CLI command paths exactly (space-delimited), so `container rm-all` in
// a config does the same thing as `dockertoolbox container rm-all`.
//
// Verbs act on all resources: selectors are a CLI concern, not a config one.
var verbs = map[string]verb{
	"overview": func(ctx context.Context, e engine.API, f ops.Flags) error { return ops.Overview(ctx, e) },
	"container ls": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.ContainerList(ctx, e, engine.Selector{})
	},
	"container stop-all": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.ContainerStopAll(ctx, e, engine.Selector{}, f)
	},
	"container rm-all": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.ContainerRemoveAll(ctx, e, engine.Selector{}, f)
	},
	"container rm-stopped": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.ContainerRemoveStopped(ctx, e, engine.Selector{}, f)
	},
	"image ls": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.ImageList(ctx, e, engine.Selector{})
	},
	"image rm-all": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.ImageRemoveAll(ctx, e, engine.Selector{}, f)
	},
	"image clean": func(ctx context.Context, e engine.API, f ops.Flags) error { return ops.ImageClean(ctx, e, f) },
	"volume ls": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.VolumeList(ctx, e, engine.Selector{})
	},
	"volume clean": func(ctx context.Context, e engine.API, f ops.Flags) error { return ops.VolumeClean(ctx, e, f) },
	"network ls": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.NetworkList(ctx, e, engine.Selector{})
	},
	"network clean": func(ctx context.Context, e engine.API, f ops.Flags) error { return ops.NetworkClean(ctx, e, f) },
	"system prune":  func(ctx context.Context, e engine.API, f ops.Flags) error { return ops.SystemPrune(ctx, e, f, false) },
	"system prune --volumes": func(ctx context.Context, e engine.API, f ops.Flags) error {
		return ops.SystemPrune(ctx, e, f, true)
	},
	"system df":    func(ctx context.Context, e engine.API, f ops.Flags) error { return ops.SystemDF(ctx, e) },
	"system reset": func(ctx context.Context, e engine.API, f ops.Flags) error { return ops.SystemReset(ctx, e, f) },
}

// lookupVerb returns the native verb for a step line, if one exists.
func lookupVerb(line string) (verb, bool) {
	v, ok := verbs[line]
	return v, ok
}

// IsVerb reports whether line names a known native verb. Used by callers
// that want to warn on a likely-mistyped verb.
func IsVerb(line string) bool {
	_, ok := verbs[line]
	return ok
}
