package main

import (
	"crm-redigo-cmd/command"
	"crm-redigo-cmd/pkg/logger"
	"github.com/spf13/cobra"
	"strconv"
)

func newCLI() *cobra.Command {
	cli := &cobra.Command{
		Use:   "crm-redigo-cmd",
		Short: "redis command for migration",
	}

	cli.AddCommand(migrateCmd())
	cli.AddCommand(validateCmd())
	cli.AddCommand(retryCmd())

	return cli
}

func migrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate all redis keys",
		Run: func(_ *cobra.Command, _ []string) {
			logger.Infof("Start migrating keys ...")
			command.Migrate()
			logger.Infof("Finish migrating keys ...")
		},
	}
}

func validateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "validate",
		Short: "Validate keys (format: validate <max keys>)",
		Run: func(_ *cobra.Command, args []string) {
			logger.Infof("Start validating keys ...")
			maxKeys := 10000
			forceRewrite := false
			if len(args) == 2 {
				threshold, err := strconv.Atoi(args[0])
				if err != nil {
					logger.Fatalf("error converting threshold")
				}
				maxKeys = threshold

				force, err := strconv.ParseBool(args[1])
				if err != nil {
					logger.Fatalf("error converting boolean")
				}
				forceRewrite = force
			}
			command.Validate(maxKeys, forceRewrite)
			logger.Infof("Finish validating keys ...")
		},
	}
}

func retryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "retry",
		Short: "Retry failed keys",
		Run: func(_ *cobra.Command, _ []string) {
			logger.Infof("Start retrying keys ...")
			command.Retry()
			logger.Infof("Finish retrying keys ...")
		},
	}
}
