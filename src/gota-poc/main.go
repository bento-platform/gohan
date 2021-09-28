package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
)

func main() {

	client := http.DefaultClient

	ucscUrl := "https://genome.ucsc.edu/cgi-bin/hgTables"

	// make initial call to get hgid
	req, err := client.Get(ucscUrl)
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
	fmt.Printf("%+v", req)
	// TODO: get Origin-Trial from header
	var originTrial string
	for key, value := range req.Header {
		if key == "Origin-Trial" {
			fmt.Println("Got 'Origin-Trial' Header")
			originTrial = value[0]
		}
	}
	if originTrial == "" {
		log.Fatal("Missing originTrial")
	}

	// TODO: get hguid from cookie
	var hguid string
	for _, cookie := range req.Cookies() {
		if cookie.Name == "hguid" {
			fmt.Println("Got 'hguid' Cookie")
			hguid = cookie.Value
		}
	}
	if hguid == "" {
		log.Fatal("Missing hguid")
	}

	var dbWg sync.WaitGroup
	allDBs := []string{"hg38", "hg19", "hg18", "hg17", "hg16"}
	for _, db := range allDBs {
		dbWg.Add(1)

		go func(_db string, _wg *sync.WaitGroup) {
			defer _wg.Done()
			fmt.Printf("Setting up %s..\n", _db)

			// begin mining
			v := url.Values{}
			// v.Add("hgsid", )
			v.Add("jsh_pageVertPos", "0")
			v.Add("clade", "mammal")
			v.Add("org", "Human")
			v.Add("db", _db)
			v.Add("hgta_group", "genes")
			v.Add("hgta_track", "knownGene")
			v.Add("hgta_table", "knownGene")
			v.Add("hgta_regionType", "genome")
			// v.Add("position", "chrM:5,904-7,445")
			v.Add("hgta_outputType", "primaryTable")
			v.Add("boolshad.sendToGalaxy", "0")
			v.Add("boolshad.sendToGreat", "0")
			v.Add("hgta_outFileName", "")
			v.Add("hgta_compressType", "none")
			v.Add("hgta_doTopSubmit", "get output")

			miningReq, _ := http.NewRequest("GET", ucscUrl, nil)
			miningReq.Header.Set("Cookie", fmt.Sprintf("hguid=%s", hguid))
			miningReq.Header.Set("Host", "genome.ucsc.edu")
			miningReq.Header.Set("Origin", "https://genome.ucsc.edu")
			miningReq.Header.Set("Referer", ucscUrl)

			fmt.Printf("Downloading %s..\n", _db)
			req, err = client.PostForm(ucscUrl, v)
			if err != nil {
				fmt.Printf("err:%v\n", err)
			}
			fmt.Printf("%+v", req)
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Printf("Error reading body: %v", err)
				return
			}

			var file *os.File
			dbPath := fmt.Sprintf("%s.csv", _db)
			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
				file, err = os.Create(dbPath)
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				file, err = os.OpenFile(dbPath, os.O_RDWR, 0755)
				if err != nil {
					log.Fatal(err)
				}
			}
			defer file.Close()

			_, err = file.WriteString(string(body))
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("Save to file %s\n", dbPath)
		}(db, &dbWg)
	}
	dbWg.Wait()
}
