package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type SessionsRunner interface {
	New(ctx context.Context, dir, prompt string) (string, error)
	Attach(id string) error
	Stop(ctx context.Context, id string) error
	Remove(ctx context.Context, id string) error
	Resume(ctx context.Context, dir, sessionID string) (string, error)
}

func NewSessionCLIs(load Loader, services Services) []*cobra.Command {
	sessions := func() (SessionsRunner, error) {
		cfg, err := load()
		if err != nil {
			return nil, err
		}
		return services.Sessions(cfg), nil
	}
	newCmd := &cobra.Command{
		Use:   "new <dir> [prompt...]",
		Short: "Start a background session in <dir> and print its job id",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := sessions()
			if err != nil {
				return err
			}
			id, err := s.New(cmd.Context(), args[0], strings.Join(args[1:], " "))
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
	attach := &cobra.Command{
		Use:   "attach <id>",
		Short: "Open a terminal on a background session",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			s, err := sessions()
			if err != nil {
				return err
			}
			return s.Attach(args[0])
		},
	}
	stop := &cobra.Command{
		Use:   "stop <id>",
		Short: "Stop a background session; its conversation is kept",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := sessions()
			if err != nil {
				return err
			}
			return s.Stop(cmd.Context(), args[0])
		},
	}
	rm := &cobra.Command{
		Use:   "rm <id>",
		Short: "Delete a background session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := sessions()
			if err != nil {
				return err
			}
			return s.Remove(cmd.Context(), args[0])
		},
	}
	resume := &cobra.Command{
		Use:   "resume <session-id>",
		Short: "Resume a parked session in the background and attach to it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := sessions()
			if err != nil {
				return err
			}
			dir, _ := cmd.Flags().GetString("dir")
			id, err := s.Resume(cmd.Context(), dir, args[0])
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), id)
			return nil
		},
	}
	resume.Flags().String("dir", "", "directory to resume in (default: the session's own cwd)")
	return []*cobra.Command{newCmd, attach, stop, rm, resume}
}
