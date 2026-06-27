package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yusuke0610/pGo-snap-cleaner/internal/decision"
	"github.com/yusuke0610/pGo-snap-cleaner/internal/findertag"
	"github.com/yusuke0610/pGo-snap-cleaner/internal/scan"
)

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:     "pGo-snap-cleaner",
		Version: version,
		Short:   "Tag Pokémon GO AR snapshots with the Finder tag pGo (red)",
		Long: "pGo-snap-cleaner scans a photo library for Pokémon GO AR snapshots and " +
			"marks them with the Finder tag \"pGo\" (red).\n" +
			"It never deletes files — you review the red tag in Finder and delete manually.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(newScanCmd(), newTagCmd(), newUntagCmd())
	return root
}

func newScanCmd() *cobra.Command {
	var recursive bool
	cmd := &cobra.Command{
		Use:   "scan <dir>",
		Short: "Walk a directory and report Pokémon GO snapshots (no changes made)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := decision.New()
			var total, matched, highCount, lowCount int
			err := scan.Walk(args[0], recursive, func(path string) error {
				total++
				out, err := c.Classify(path)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "  skip %s: %v\n", path, err)
					return nil
				}
				if out.Result.Matched {
					matched++
					switch out.Confidence {
					case decision.High:
						highCount++
					case decision.Low:
						lowCount++
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  [%s] %s\n", out.Confidence, path)
				}
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"\nscanned %d image(s): %d Pokémon GO snapshot(s) (%d high, %d low confidence)\n",
				total, matched, highCount, lowCount)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "descend into subdirectories")
	return cmd
}

func newTagCmd() *cobra.Command {
	var recursive bool
	cmd := &cobra.Command{
		Use:   "tag <dir>",
		Short: "Add the pGo (red) Finder tag to detected snapshots (idempotent)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c := decision.New()
			var tagged, already int
			err := scan.Walk(args[0], recursive, func(path string) error {
				out, err := c.Classify(path)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "  skip %s: %v\n", path, err)
					return nil
				}
				if !out.ShouldTag() {
					return nil
				}
				added, err := findertag.AddPGo(path)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "  failed %s: %v\n", path, err)
					return nil
				}
				if added {
					tagged++
					fmt.Fprintf(cmd.OutOrStdout(), "  tagged %s\n", path)
				} else {
					already++
				}
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(),
				"\ntagged %d file(s); %d already had the pGo tag\n", tagged, already)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "descend into subdirectories")
	return cmd
}

func newUntagCmd() *cobra.Command {
	var recursive bool
	cmd := &cobra.Command{
		Use:   "untag <dir>",
		Short: "Remove only the pGo Finder tag, leaving other tags untouched",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var removed int
			err := scan.Walk(args[0], recursive, func(path string) error {
				ok, err := findertag.RemovePGo(path)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "  failed %s: %v\n", path, err)
					return nil
				}
				if ok {
					removed++
					fmt.Fprintf(cmd.OutOrStdout(), "  untagged %s\n", path)
				}
				return nil
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nremoved the pGo tag from %d file(s)\n", removed)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "descend into subdirectories")
	return cmd
}
