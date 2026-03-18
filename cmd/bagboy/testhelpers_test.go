package main

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// resetAllFlags resets every flag on cmd and its entire subcommand tree to its
// default value and marks it as unchanged.  Call this before rootCmd.Execute()
// in tests to prevent Cobra/pflag flag state from leaking between test cases
// (pflag does not reset flag values between successive Parse calls).
func resetAllFlags(cmd *cobra.Command) {
	reset := func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	}
	cmd.Flags().VisitAll(reset)
	cmd.PersistentFlags().VisitAll(reset)
	for _, sub := range cmd.Commands() {
		resetAllFlags(sub)
	}
}
