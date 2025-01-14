package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "githelper",
	Short: "A CLI tool to simplify complex GitHub workflows",
	Long: `GitHelper is a command-line tool that simplifies complex GitHub workflows
that are not straightforward with basic Git commands. It provides various
utilities to manage repositories, branches, and common Git operations.`,
}

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.githelper.yaml)")
}

func initConfig() {
	debug := rootCmd.PersistentFlags().Lookup("debug").Value.String() == "true"

	// Always show config file location in debug mode
	if debug {
		if cfgFile := rootCmd.PersistentFlags().Lookup("config").Value.String(); cfgFile != "" {
			fmt.Printf("Using config file specified by flag: %s\n", cfgFile)
		} else {
			home, _ := os.UserHomeDir()
			fmt.Printf("No config file specified, will look in: %s/.githelper.yaml\n", home)
		}
	}

	// Debug flag check should be first
	debug = rootCmd.PersistentFlags().Lookup("debug").Value.String() == "true"

	if cfgFile := rootCmd.PersistentFlags().Lookup("config").Value.String(); cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".githelper"
		viper.AddConfigPath(home)
		viper.SetConfigName(".githelper")
		viper.SetConfigType("yaml")
		
		// Add debug line to show where we're looking
		if debug {
			fmt.Printf("Looking for config file at: %s/.githelper.yaml\n", home)
			if _, err := os.Stat(fmt.Sprintf("%s/.githelper.yaml", home)); err != nil {
				fmt.Printf("Config file status: %v\n", err)
			} else {
				fmt.Println("Config file exists")
			}
		}
	}

	// Read environment variables
	viper.SetEnvPrefix("GITHELPER")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if debug {
				fmt.Println("No config file found")
			}
		} else {
			fmt.Fprintln(os.Stderr, "Error reading config file:", err)
			os.Exit(1)
		}
	}

	if debug {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		fmt.Printf("GitHub token present: %v\n", viper.GetString("github_token") != "")
		fmt.Printf("OpenAI API key present: %v\n", viper.GetString("openai_api_key") != "")
		fmt.Printf("Config values: %+v\n", viper.AllSettings())
	}

	if debug {
		fmt.Printf("Final config state:\n")
		fmt.Printf("Config file used: %s\n", viper.ConfigFileUsed())
		fmt.Printf("All settings: %#v\n", viper.AllSettings())
		fmt.Printf("GitHub token length: %d\n", len(viper.GetString("github_token")))
	}
}