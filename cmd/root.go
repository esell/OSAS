package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
)

type conf struct {
	DBDSN       string `json:"dbDSN"`
	WordlistURL string `json:"wordlistURL"`
}

var (
	configFile   string
	parsedConfig = conf{}
	rootCmd      = &cobra.Command{
		Use:   "osas-reborn",
		Short: "A tool to scan for open Azure storage account blob containers",
		Long:  `A tool to scan for open Azure storage account blob containers. Additionally, it supports listing and downloading files from open containers.`,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "conf.json", "config file")
}

func initConfig() {
	confFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("unable to read config file!")
		os.Exit(1)
	}
	if err := json.Unmarshal(confFile, &parsedConfig); err != nil {
		fmt.Println("unable to marshal config file!")
		os.Exit(1)
	}

}
