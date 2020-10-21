package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type openContainer struct {
	URL       string    `json:"url"`
	LastCheck time.Time `json:"last_check"`
}

var db *sql.DB

func main() {
	var err error
	db, err = sql.Open("mysql", "username:password@tcp(dbhost:dbport)/dbname?parseTime=true")

	if err != nil {
		panic(err)
	}

	defer db.Close()

	http.HandleFunc("/", indexHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getOpenContainers() []openContainer {
	demOpen := make([]openContainer, 0)
	rows, err := db.Query("SELECT * FROM open_containers ORDER BY last_check DESC")
	if err != nil {
		return demOpen
	}

	for rows.Next() {
		tempOc := openContainer{}
		err = rows.Scan(&tempOc.URL, &tempOc.LastCheck)
		if err != nil {
			fmt.Println(err)
		}
		demOpen = append(demOpen, tempOc)
	}
	return demOpen
}

func getNotScannedCount() int {
	var notScannedCount int
	err := db.QueryRow("SELECT count(*) FROM stg_accts WHERE last_check = '2020-07-24 21:34:27'").Scan(&notScannedCount)
	if err != nil {
		return 0
	}

	return notScannedCount
}

func getScannedCount() int {
	var scannedCount int
	err := db.QueryRow("SELECT count(*) FROM stg_accts WHERE last_check != '2020-07-24 21:34:27'").Scan(&scannedCount)
	if err != nil {
		return 0
	}

	return scannedCount
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	topText := "<html><head><title>hola</title></head><body>"
	countText := "<b>To Scan:</b> " + strconv.Itoa(getNotScannedCount()) + "<br><b>Scanned:</b> " + strconv.Itoa(getScannedCount()) + "<br><br><br>"

	var openContainerText strings.Builder
	bottomText := "</body></html>"

	// ROFL!
	oc := getOpenContainers()
	openContainerText.WriteString("<table><tr><th>Container</th><th>Last Check</th></tr>")

	for _, v := range oc {
		openContainerText.WriteString("<tr><td>" + v.URL + "</td><td>" + v.LastCheck.String() + "</td></tr>")
	}

	openContainerText.WriteString("</table>")
	fmt.Fprintf(w, topText+countText+openContainerText.String()+bottomText)
}
