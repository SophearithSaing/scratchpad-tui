package main

import (
	"fmt"
	"io"
	"os"

	tea "charm.land/bubbletea/v2"
)

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "scratchpad: %v\n", err)
		os.Exit(1)
	}
}

func run(in io.Reader, out io.Writer) error {
	store, err := newSessionStore()
	if err != nil {
		return err
	}

	session, err := store.load()
	if err != nil {
		return err
	}

	appModel := newAppModel(store, session)
	program := tea.NewProgram(appModel, tea.WithInput(in), tea.WithOutput(out))
	_, err = program.Run()
	return err
}
