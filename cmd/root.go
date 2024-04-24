package cmd

import (
	"ibanking-scraper/config"
	"ibanking-scraper/internal/logger"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type RootCommand struct {
	cobraCommand *cobra.Command

	configFile string
}

var rootCommand = RootCommand{
	cobraCommand: &cobra.Command{
		Use: "mpn",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	NewCmdRoot()
}

func NewCmdRoot() *cobra.Command {
	cmd := rootCommand.cobraCommand

	cmd.PersistentFlags().StringVarP(&rootCommand.configFile, "config", "c", "", "MPN API configuration file")
	viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))

	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		cmd.SilenceUsage = true

		config.Load()
		logger.Setup()
	}

	cmd.AddCommand(NewCmdServer())
	cmd.AddCommand(NewCmdVersion())
	return cmd
}

func initConfig() {
	cfg := viper.GetString("config")
	if cfg != "" {
		viper.SetConfigFile(cfg)
		viper.SetConfigType("yaml")
	}

	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file: ", viper.ConfigFileUsed())
	}
}

func Execute() {
	if err := rootCommand.cobraCommand.Execute(); err != nil {
		os.Exit(1)
	}
}
