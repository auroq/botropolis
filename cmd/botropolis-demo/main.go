package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/auroq/botropolis/pkg/demo"
	"github.com/auroq/botropolis/pkg/version"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, out, errOut io.Writer) int {
	root := &cobra.Command{
		Use:           "botropolis-demo",
		Short:         "Stage a demo city from the recorded corpus, so no real session is ever filmed",
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetVersionTemplate("{{.Name}} {{.Version}}\n")
	root.AddCommand(stageCommand(), unstageCommand(), filmCommand())
	root.SetOut(out)
	root.SetErr(errOut)
	root.SetArgs(args)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(errOut, "botropolis-demo: %v\n", err)
		return 1
	}
	return 0
}

func stageCommand() *cobra.Command {
	var corpus, scenarioPath, home string
	var hold time.Duration
	cmd := &cobra.Command{
		Use:   "stage",
		Short: "Write a home for a scenario and start stand-ins for its live sessions",
		Long: "Writes a Claude home for the scenario into --home: each placed session's\n" +
			"transcript moved to end relative to now, a session record for every one\n" +
			"that is not parked, pointed at a detached sleep, and the MCP config.\n" +
			"Point botropolis at it with --home and --socket, and run unstage after.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			scenario, err := demo.LoadScenario(scenarioPath)
			if err != nil {
				return err
			}
			if scenario.Name == "" {
				scenario.Name = filepath.Base(scenarioPath)
			}
			if err := demo.Unstage(home); err != nil {
				return err
			}
			if err := os.RemoveAll(filepath.Join(home, ".claude")); err != nil {
				return err
			}
			if err := os.MkdirAll(home, 0o700); err != nil {
				return err
			}
			if err := demo.Stage(demo.Corpus{Dir: corpus}, scenario, home, time.Now(), demo.Sleepers(home, hold)); err != nil {
				_ = demo.Unstage(home)
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "staged %s: %d sessions in %s\n", scenario.Name, len(scenario.Sessions), home)
			return nil
		},
	}
	cmd.Flags().StringVar(&corpus, "corpus", "demo/corpus", "directory of recorded transcripts")
	cmd.Flags().StringVar(&scenarioPath, "scenario", "", "scenario file")
	cmd.Flags().StringVar(&home, "home", "", "directory to stage the home into")
	cmd.Flags().DurationVar(&hold, "hold", time.Hour, "how long a stand-in lives if never unstaged")
	_ = cmd.MarkFlagRequired("scenario")
	_ = cmd.MarkFlagRequired("home")
	return cmd
}

func unstageCommand() *cobra.Command {
	var home string
	cmd := &cobra.Command{
		Use:   "unstage",
		Short: "End a staged home's stand-ins",
		Args:  cobra.NoArgs,
		RunE: func(*cobra.Command, []string) error {
			return demo.Unstage(home)
		},
	}
	cmd.Flags().StringVar(&home, "home", "", "the staged home")
	_ = cmd.MarkFlagRequired("home")
	return cmd
}

func filmCommand() *cobra.Command {
	var corpus, out, botropolis, ffmpeg string
	cmd := &cobra.Command{
		Use:   "film SCENARIO...",
		Short: "Make every shot of each scenario and write a manifest",
		Long: "Stages each scenario afresh for every shot, runs the city headless\n" +
			"against it with HOME and the XDG directories in a scratch directory,\n" +
			"encodes clips with ffmpeg, and writes OUT/manifest.json.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, paths []string) error {
			if botropolis == "" {
				self, err := os.Executable()
				if err != nil {
					return err
				}
				botropolis = filepath.Join(filepath.Dir(self), "botropolis")
			}
			f := demo.Filmer{
				Corpus: demo.Corpus{Dir: corpus}, Botropolis: botropolis, FFmpeg: ffmpeg,
				Out: out, Version: version.Version, Log: cmd.ErrOrStderr(), Now: time.Now,
			}
			manifest, err := f.Film(paths)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "filmed %d files into %s\n", len(manifest.Media), out)
			return nil
		},
	}
	cmd.Flags().StringVar(&corpus, "corpus", "demo/corpus", "directory of recorded transcripts")
	cmd.Flags().StringVar(&out, "out", "dist/demo", "directory to write media and manifest.json into")
	cmd.Flags().StringVar(&botropolis, "botropolis", "", "the botropolis binary (default: beside this one)")
	cmd.Flags().StringVar(&ffmpeg, "ffmpeg", "ffmpeg", "the ffmpeg binary")
	return cmd
}
