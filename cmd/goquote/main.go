package main

import (
	"goquotebot/pkg/config"

	"github.com/spf13/cobra"
)

func main() {
	config := &config.Config{}

	cmd := cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			err := config.RegisterConfigFile()
			if err != nil {
				return err
			}

			err = run(config)
			return err
		},
	}

	config.RegisterFlags(cmd.PersistentFlags())

	err := cmd.Execute()
	if err != nil {
		panic(err)
	}
}
