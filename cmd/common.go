package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// confirm prompts the user with a yes/no question on stdin.
//
// It returns true when the user answers "y" or "yes" (case-insensitive) and
// false otherwise. An empty answer (just pressing Enter) is treated as "no",
// matching the conventional "y/N" semantics used throughout the CLI.
//
// When stdin is not a terminal (e.g. piped input, EOF immediately) the answer
// is treated as "no" so that destructive operations never run unattended.
func confirm(prompt string) bool {
	fmt.Printf("%s [y/N]: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	answer, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	answer = strings.TrimSpace(strings.ToLower(answer))
	return answer == "y" || answer == "yes"
}

// printErr writes a styled error message to stderr.
func printErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}
