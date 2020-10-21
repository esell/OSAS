package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type conf struct {
	DBDSN       string `json:"dbDSN"`
	WordlistURL string `json:"wordlistURL"`
}

type stgAcct struct {
	URL       string    `json:"url"`
	LastCheck time.Time `json:"last_check"`
	IsOpen    bool      `json:"is_open"`
}

var (
	configFile   = flag.String("c", "conf.json", "config file location")
	workerCount  = flag.Int("w", 10, "worker count")
	maxConns     = flag.Int("m", 100, "max connections to host")
	db           *sql.DB
	parsedConfig = conf{}
)

func main() {
	flag.Parse()

	confFile, err := ioutil.ReadFile(*configFile)
	if err != nil {
		fmt.Println("unable to read config file!")
		os.Exit(1)
	}
	if err := json.Unmarshal(confFile, &parsedConfig); err != nil {
		fmt.Println("unable to marshal config file!")
		os.Exit(1)
	}

	var wg sync.WaitGroup
	jobs := make(chan string, *workerCount*1000)

	tr := &http.Transport{
		MaxIdleConns:        *maxConns,
		MaxIdleConnsPerHost: *maxConns,
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	fmt.Println("downloading wordlist: ", parsedConfig.WordlistURL)
	response, err := http.Get(parsedConfig.WordlistURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	file, err := os.Create("word.list")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = io.Copy(file, response.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	response.Body.Close()
	file.Close()

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

	for w := 1; w <= *workerCount; w++ {
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

	go scanLines("word.list", jobs)

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

	resp, err := http.Get("https://" + host + "/" + dir + "?restype=container&comp=list")
	if err != nil {
		return err
	}

	// in order to re-use connections we have
	// to read the body (https://golang.org/pkg/net/http/#Client.Do)
	ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 {
		//winner
		fmt.Println("SUCCESS: https://" + host + "/" + dir + "?restype=container&comp=list")
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

		_, err = insertRes.Exec("https://"+host+"/"+dir, time.Now())
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

	for scanner.Scan() {
		blah := scanner.Text()
		results <- blah
		recordCount++
	}

	close(results)
}
