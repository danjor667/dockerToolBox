package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dockerToolBox/internal/ui"
)

// confirmDestructive enforces the safety contract every destructive
// command must use: it lists the affected items, then gates the action
// behind --dry-run and confirmation. It returns true only if the caller
// should proceed.
//
//   - no targets        -> prints a notice, returns false
//   - --dry-run         -> prints the targets, returns false
//   - --yes             -> returns true
//   - otherwise         -> interactive y/N prompt (default No)
func confirmDestructive(action string, targets []string) (bool, error) {
	if len(targets) == 0 {
		ui.Infof("nothing to %s", action)
		return false, nil
	}

	ui.Header("%s — %d item(s):", action, len(targets))
	for _, t := range targets {
		ui.Step("%s", t)
	}

	if flagDryRun {
		ui.Infof("dry-run: no changes made")
		return false, nil
	}
	if flagYes {
		return true, nil
	}
	return prompt(fmt.Sprintf("Proceed to %s these %d item(s)?", action, len(targets)))
}

// prompt asks a yes/no question. Anything other than y/yes (including
// EOF or no input) is treated as No.
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
