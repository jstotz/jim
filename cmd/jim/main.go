package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jstotz/jim/internal/jim"
)

func main() {
	// Phase 1 goals:
	// ##############
	// * Read text from a file
	// * Render to the screen
	// * Allow scrolling the viewport up and down
	// * Allow moving the cursor to specific positions

	if len(os.Args) != 2 {
		log.Fatalln("must specify file path")
	}

	filepath := os.Args[1]
	if err := editFile(filepath); err != nil {
		log.Fatalln("failed to edit file:", err)
	}
}

func editFile(path string) error {
	logFile, err := os.Create("jim.log")
	if err != nil {
		return err
	}
	e := jim.NewEditor(nil, nil, logFile)
	if err := e.Setup(); err != nil {
		return fmt.Errorf("editor setup: %w", err)
	}
	if err := e.LoadFile(path); err != nil {
		return fmt.Errorf("editor edit file: %w", err)
	}
	if err := e.Start(); err != nil {
		return fmt.Errorf("editor exited: %w", err)
	}
	return nil
}
