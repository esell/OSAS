package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"text/tabwriter"
)

type EnumerationResults struct {
	XMLName       xml.Name `xml:"EnumerationResults"`
	Text          string   `xml:",chardata"`
	ContainerName string   `xml:"ContainerName,attr"`
	Blobs         struct {
		Text string `xml:",chardata"`
		Blob []struct {
			Text       string `xml:",chardata"`
			Name       string `xml:"Name"`
			URL        string `xml:"Url"`
			Properties struct {
				Text            string `xml:",chardata"`
				LastModified    string `xml:"Last-Modified"`
				Etag            string `xml:"Etag"`
				ContentLength   string `xml:"Content-Length"`
				ContentType     string `xml:"Content-Type"`
				ContentEncoding string `xml:"Content-Encoding"`
				ContentLanguage string `xml:"Content-Language"`
				ContentMD5      string `xml:"Content-MD5"`
				CacheControl    string `xml:"Cache-Control"`
				BlobType        string `xml:"BlobType"`
				LeaseStatus     string `xml:"LeaseStatus"`
			} `xml:"Properties"`
		} `xml:"Blob"`
	} `xml:"Blobs"`
	NextMarker string `xml:"NextMarker"`
}

var (
	saurl      = flag.String("u", "https://blah.blob.core.windows.net/blah", "URL")
	doDownload = flag.Bool("d", false, "download blobs")
)

func main() {
	flag.Parse()

	if *saurl == "https://blah.blob.core.windows.net/blah" {
		fmt.Println("dur")
		os.Exit(1)
	}

	var sa EnumerationResults
	response, err := http.Get(*saurl + "?restype=container&comp=list")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer response.Body.Close()

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		//
	}

	xml.Unmarshal(contents, &sa)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	defer w.Flush()

	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", "Name", "LastModified", "ContentLength", "URL")
	fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", "--------------", "--------------", "--------------", "--------------")
	for _, b := range sa.Blobs.Blob {
		if b.URL == "" {
			fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", b.Name, b.Properties.LastModified, b.Properties.ContentLength, *saurl+"/"+b.Name)
		} else {
			fmt.Fprintf(w, "\n %s\t%s\t%s\t%s\t", b.Name, b.Properties.LastModified, b.Properties.ContentLength, b.URL)
		}
	}
	fmt.Fprintf(w, "\n\n")

	if *doDownload {
		for _, b := range sa.Blobs.Blob {
			if b.URL == "" {
				downloadBlob(b.Name, *saurl+"/"+b.Name)
			} else {
				downloadBlob(b.Name, b.URL)
			}
		}
	}
}

func downloadBlob(blobName, blobURL string) {
	fmt.Println("downloading: ", blobURL)
	response, err := http.Get(blobURL)
	if err != nil {
		fmt.Println(err)
	}

	var dir, fileName string
	var localFile *os.File

	dir, fileName = filepath.Split(blobName)

	if dir == "" {
		localFile, err = os.Create(fileName)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			fmt.Println(err)
		}
		localFile, err = os.Create(filepath.Join(dir, fileName))
		if err != nil {
			fmt.Println(err)
		}
	}

	_, err = io.Copy(localFile, response.Body)
	if err != nil {
		fmt.Println(err)
	}

	response.Body.Close()
	localFile.Close()
}
