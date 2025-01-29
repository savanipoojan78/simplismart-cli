package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

func GenerateDocs(rootCmd *cobra.Command) {
	docsCmd := &cobra.Command{
		Use:          "docs",
		Short:        "Generate CLI docs in various formats",
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			err = os.MkdirAll("docs", os.ModePerm)
			if err != nil {
				return errors.Wrap(err, "unable to create directory")
			}
			err = doc.GenMarkdownTree(rootCmd, "docs")
			if err != nil {
				return fmt.Errorf("generating docs failed")
			}
			return nil
		},
	}
	rootCmd.AddCommand(docsCmd)
}
