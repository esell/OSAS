package cmd

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
)

type stgAcct struct {
	URL       string    `json:"url"`
	LastCheck time.Time `json:"last_check"`
	IsOpen    bool      `json:"is_open"`
}

var (
	workerCount  int
	maxConns     int
	scanWordList string
	db           *sql.DB
	scanCmd      = &cobra.Command{
		Use:   "scan",
		Short: "scan a storage account for open blob containers",
		Long:  `scan a storage account for open blob containers`,
		Run:   runScan,
	}
)

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&scanWordList, "wordlist", "w", "words.list", "Word list to use")
	scanCmd.MarkFlagRequired("wordlist")
	scanCmd.Flags().IntVarP(&workerCount, "count", "t", 10, "worker count")
	scanCmd.Flags().IntVarP(&maxConns, "connections", "m", 100, "max connections to host")
}

func runScan(cmd *cobra.Command, args []string) {
	confFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Println("unable to read config file!")
		os.Exit(1)
	}
	if err := json.Unmarshal(confFile, &parsedConfig); err != nil {
		fmt.Println("unable to marshal config file!")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	jobs := make(chan string, workerCount*1000)

	tr := &http.Transport{
		MaxIdleConns:        maxConns,
		MaxIdleConnsPerHost: maxConns,
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	db, err = sql.Open("mysql", parsedConfig.DBDSN)
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

	var stgAcctTemp stgAcct
	//TODO: avoid two scanners getting same SA?
	err = db.QueryRow("SELECT url FROM stg_accts ORDER BY last_check ASC LIMIT 1").Scan(&stgAcctTemp.URL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("scanning: ", stgAcctTemp.URL)

	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go worker(stgAcctTemp.URL, jobs, &wg, w, client)
	}

	updateRes, err := db.Prepare("UPDATE stg_accts SET last_check=? WHERE url=?")
	if err != nil {
		fmt.Println(err)
	}
	defer updateRes.Close()

	_, err = updateRes.Exec(time.Now(), stgAcctTemp.URL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go scanLines(scanWordList, jobs)

	wg.Wait()

}

func worker(host string, jobs <-chan string, wg *sync.WaitGroup, workerID int, client *http.Client) {
	for {
		job, more := <-jobs
		if more {
			err := getContainer(host, job, client)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			wg.Done()
			return
		}
	}
}

func getContainer(host, dir string, client *http.Client) error {

	// try lookup to avoid wasting time
	_, err := net.LookupHost(host + ".blob.core.windows.net")
	if err != nil {
		fmt.Println("DNS lookup failed, exiting")
		os.Exit(1)
	}
	resp, err := http.Get("https://" + host + ".blob.core.windows.net/" + dir + "?restype=container&comp=list")
	if err != nil {
		return err
	}

	// in order to re-use connections we have
	// to read the body (https://golang.org/pkg/net/http/#Client.Do)
	ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		//winner
		fmt.Println("SUCCESS: https://" + host + ".blob.core.windows.net/" + dir + "?restype=container&comp=list")
		if err != nil {
			fmt.Println(err)
			return err
		}

		updateRes, err := db.Prepare("UPDATE stg_accts SET is_open=true WHERE url=?")
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer updateRes.Close()

		_, err = updateRes.Exec(host)
		if err != nil {
			fmt.Println(err)
			return err
		}

		// store full URL
		insertRes, err := db.Prepare("INSERT INTO open_containers(url, last_check) VALUES(?,?) ")
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer insertRes.Close()

		_, err = insertRes.Exec("https://"+host+".blob.core.windows.net/"+dir, time.Now())
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	resp.Body.Close()

	return nil
}

func scanLines(path string, results chan<- string) {
	recordCount := 0
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanWords)

	// container names can only have lowercase, numbers and a dash
	var isValidContainerName = regexp.MustCompile(`^[a-z0-9-]*$`).MatchString

	for scanner.Scan() {
		blah := scanner.Text()
		// container names must be 3-63 chars
		if len(blah) >= 3 && len(blah) <= 63 && isValidContainerName(blah) {
			results <- blah
			recordCount++
		}
	}

	close(results)
}
