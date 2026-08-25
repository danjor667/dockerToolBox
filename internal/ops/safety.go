package ops

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dockerToolBox/internal/ui"
)

// Flags carries the safety toggles that every operation honors. They are
// passed explicitly (rather than read from a global) so ops has no
// dependency on the command layer and stays testable.
type Flags struct {
	DryRun bool
	Yes    bool
}

// confirmDestructive enforces the safety contract every destructive
// operation must use: it lists the affected items, then gates the action
// behind DryRun and confirmation. It returns true only if the caller
// should proceed.
//
//   - no targets  -> prints a notice, returns false
//   - DryRun      -> prints the targets, returns false
//   - Yes         -> returns true
//   - otherwise   -> interactive y/N prompt (default No)
func confirmDestructive(action string, targets []string, f Flags) (bool, error) {
	if len(targets) == 0 {
		ui.Infof("nothing to %s", action)
		return false, nil
	}

	ui.Header("%s — %d item(s):", action, len(targets))
	for _, t := range targets {
		ui.Step("%s", t)
	}

	if f.DryRun {
		ui.Infof("dry-run: no changes made")
		return false, nil
	}
	if f.Yes {
		return true, nil
	}
	return prompt(fmt.Sprintf("Proceed to %s these %d item(s)?", action, len(targets)))
}

// confirmPrune gates a prune-style action (no explicit target list) behind
// DryRun and confirmation. dryMsg is shown under DryRun; question is the
// y/N prompt. It returns true only if the caller should proceed.
func confirmPrune(dryMsg, question string, f Flags) (bool, error) {
	if f.DryRun {
		ui.Infof("dry-run: %s", dryMsg)
		return false, nil
	}
	if f.Yes {
		return true, nil
	}
	return prompt(question)
}

// prompt asks a yes/no question. Anything other than y/yes (including EOF
// or no input) is treated as No.
func prompt(question string) (bool, error) {
	fmt.Printf("%s [y/N]: ", question)
	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return false, nil
	}
	switch strings.ToLower(strings.TrimSpace(line)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}
