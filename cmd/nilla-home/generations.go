package main

import (
	"cmp"
	"context"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"time"

	"github.com/arnarg/nilla-utils/internal/generation"
	"github.com/arnarg/nilla-utils/internal/tui"
	"github.com/arnarg/nilla-utils/internal/util"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
)

func sortGenerationsDesc(generations []*generation.HomeGeneration) {
	slices.SortFunc(generations, func(a, b *generation.HomeGeneration) int {
		return cmp.Compare(b.ID, a.ID)
	})
}

func listGenerations(ctx context.Context, cmd *cli.Command) error {
	// Get current generation
	current, err := generation.CurrentHomeGeneration()
	if err != nil {
		return err
	}

	// List all generations
	generations, err := generation.ListHomeGenerations()
	if err != nil {
		return err
	}

	// Sort the list in reverse by ID
	sortGenerationsDesc(generations)

	// Build table
	headers := []string{"Generation", "Build date", "Home Manager version"}
	rows := [][]string{}
	for _, gen := range generations {
		pre := " "
		if gen.ID == current.ID {
			pre = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true).
				SetString("*").
				String()
		}

		rows = append(rows, []string{
			fmt.Sprintf("%s %d", pre, gen.ID),
			gen.BuildDate.Format(time.DateTime),
			gen.Version,
		})
	}

	fmt.Println(util.RenderTable(headers, rows...))

	return nil
}

type genAction struct {
	generation *generation.HomeGeneration
	keep       bool
}

func cleanGenerations(ctx context.Context, cmd *cli.Command) error {
	// Parse parameters
	keep := cmd.Uint("keep")
	foundCurrent := false

	// Get current generation
	current, err := generation.CurrentHomeGeneration()
	if err != nil {
		return err
	}

	// List all generations
	generations, err := generation.ListHomeGenerations()
	if err != nil {
		return err
	}

	// Sort the list in reverse by ID
	sortGenerationsDesc(generations)

	// Make a plan
	remaining := keep
	actions := []genAction{}
	for _, gen := range generations {
		doKeep := remaining > 0

		if gen.ID == current.ID {
			doKeep = true
			foundCurrent = true
		} else if !foundCurrent && remaining == 1 {
			doKeep = false
		}

		if doKeep {
			remaining -= 1
		}

		actions = append(actions, genAction{gen, doKeep})
	}

	// Build plan table
	headers := []string{"Generation", "Build date", "Home Manager version"}
	rows := [][]string{}
	keepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	delStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	for _, action := range actions {
		gen := action.generation
		pre := " "
		if gen.ID == current.ID {
			pre = lipgloss.NewStyle().
				Foreground(lipgloss.Color("13")).
				Bold(true).
				SetString("*").
				String()
		}

		var style lipgloss.Style
		if action.keep {
			style = keepStyle
		} else {
			style = delStyle
		}

		rows = append(rows, []string{
			fmt.Sprintf(
				"%s %s",
				pre,
				style.SetString(strconv.Itoa(gen.ID)).String(),
			),
			style.SetString(gen.BuildDate.Format(time.DateTime)).String(),
			style.SetString(gen.Version).String(),
		})
	}

	//
	// Display plan
	//
	printSection("Plan")
	fmt.Fprintln(os.Stderr, util.RenderTable(headers, rows...))

	//
	// Ask Confirmation
	//
	if !cmd.Bool("confirm") {
		doContinue, err := tui.RunConfirm("Do you want to continue?")
		if err != nil {
			return err
		}
		if !doContinue {
			return nil
		}
	}

	//
	// Delete generation links
	//
	for _, action := range actions {
		if !action.keep {
			if err := action.generation.Delete(); err != nil {
				return err
			}
		}
	}

	//
	// Collect garbage
	//
	fmt.Fprintln(os.Stderr)
	printSection("Collecting garbage from nix store")

	gc := exec.CommandContext(ctx, "nix", "store", "gc", "-v")
	gc.Stdout = os.Stderr
	gc.Stderr = os.Stderr

	return gc.Run()
}
