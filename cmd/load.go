package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

var (
	wordList string
	loadCmd  = &cobra.Command{
		Use:   "load",
		Short: "load a list of targets into the database",
		Long:  `load a list of targets into the database`,
		Run:   runLoad,
	}
)

func init() {
	rootCmd.AddCommand(loadCmd)
	loadCmd.Flags().StringVarP(&wordList, "wordlist", "w", "", "Word list to load")
	loadCmd.MarkFlagRequired("wordlist")
	//TODO: ??
	loadCmd.MarkPersistentFlagRequired("config")

}

func runLoad(cmd *cobra.Command, args []string) {

	db, err := sql.Open("mysql", parsedConfig.DBDSN)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = db.Ping()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()

	insertRes, err := db.Prepare("INSERT INTO stg_accts(url, last_check, is_open) VALUES(?,?,?) ")
	if err != nil {
		fmt.Println(err)
	}
	defer insertRes.Close()
	// read file
	file, err := os.Open(wordList)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	// storage account names can only have lowercase and numbers
	var isValidName = regexp.MustCompile(`^[a-z0-9]*$`).MatchString

	for scanner.Scan() {
		blah := scanner.Text()
		// storage account names need to be 3-27 characters
		if len(blah) >= 3 && len(blah) <= 27 && isValidName(blah) {
			_, err = insertRes.Exec(blah, time.Now(), false)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
