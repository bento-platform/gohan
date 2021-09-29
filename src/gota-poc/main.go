package main

import (
	"api/models"
	"api/models/constants"
	assemblyId "api/models/constants/assembly-id"
	"api/models/ingest/structs"
	"api/services"
	"api/utils"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/go-gota/gota/dataframe"
	"github.com/kelseyhightower/envconfig"
)

func main() {

	// Gather environment variables
	var cfg models.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	// TEMP: SECURITY RISK
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//

	// Service Connections:
	// -- Elasticsearch
	es := utils.CreateEsConnection(&cfg)
	iz := services.NewIngestionService(es)
	iz.Init()

	ucscUrl := "https://genome.ucsc.edu/cgi-bin/hgTables"

	// make initial call to get hgid
	client := http.DefaultClient
	req, err := client.Get(ucscUrl)
	if err != nil {
		fmt.Printf("err:%v\n", err)
	}
	fmt.Printf("%+v", req)

	// get Origin-Trial from header
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

	// get hguid from cookie
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

	assemblyIdMap := map[constants.AssemblyId]string{
		assemblyId.GRCh38: "hg38",
		assemblyId.GRCh37: "hg19",
		assemblyId.NCBI36: "hg18",
		assemblyId.NCBI35: "hg17",
		assemblyId.NCBI34: "hg16",
	}

	var dbWg sync.WaitGroup
	for _, db := range assemblyIdMap {
		dbWg.Add(1)

		go func(_db string, _wg *sync.WaitGroup) {
			defer _wg.Done()

			dbPath := fmt.Sprintf("%s.csv", _db)

			if _, err := os.Stat(dbPath); os.IsNotExist(err) {
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
			} else {
				fmt.Printf("%s already downloaded!\n", dbPath)
			}

		}(db, &dbWg)
	}
	dbWg.Wait()

	var geneWg sync.WaitGroup

	for assId, db := range assemblyIdMap {
		// Read one file at a time
		dbPath := fmt.Sprintf("%s.csv", db)

		content, err := ioutil.ReadFile(dbPath) // the file is inside the local directory
		if err != nil {
			fmt.Println("Err")
		}

		// Gota
		df := dataframe.ReadCSV(strings.NewReader(string(content)),
			dataframe.WithDelimiter('\t'),
			dataframe.HasHeader(true))
		fmt.Printf("%s :\n", dbPath)

		var (
			chromHeaderKey     = 0
			startKey           = 1
			endKey             = 2
			nameHeaderKeys     = []int{3}
			geneNameHeaderKeys []int
		)

		var columnsToPrint []string
		if assId == assemblyId.GRCh38 {
			// GRCh38 dataset has multiple name fields (name, name2) and
			// also includes gene name fields (geneName, geneName2)
			columnsToPrint = append(columnsToPrint, "#chrom", "chromStart", "chromEnd", "name", "name2", "geneName", "geneName2")
			nameHeaderKeys = append(nameHeaderKeys, 4)
			geneNameHeaderKeys = append(geneNameHeaderKeys, 5, 6)
		} else {
			columnsToPrint = append(columnsToPrint, "chrom", "txStart", "txEnd", "#name")
		}

		df = df.Select(columnsToPrint)
		fmt.Println(df)

		// discover name indexes

		for n, record := range df.Records() {
			if n == 0 {
				continue
			}

			geneWg.Add(1)
			go func(_record []string, _chromHeaderKey int,
				_startKey int, _endKey int,
				_nameHeaderKeys []int, _geneNameHeaderKeys []int,
				_assId constants.AssemblyId,
				_gwg *sync.WaitGroup) {
				// fmt.Printf("row : %s\n", row)

				// create instance of a Gene structure
				var names, geneNames []string
				var chromStart, chromEnd int

				// discover names
				for _, nk := range _nameHeaderKeys {
					names = append(names, _record[nk])
				}
				for _, nk := range geneNameHeaderKeys {
					geneNames = append(geneNames, _record[nk])
				}

				//clean chromosome
				chromosomeClean := strings.ReplaceAll(strings.ReplaceAll(_record[_chromHeaderKey], "chr", ""), "#", "")

				// skip this record if the chromosome contians "scaffolding", i.e 'chr1_something_something'
				if strings.Contains(chromosomeClean, "_") {
					geneWg.Done()
					return
				}

				// TODO: fomarmalize
				// if chromosome MT, set to 0
				// if chromosome X, set to -1
				// if chromosome Y, set to -2
				var chromosome int
				if strings.Contains(strings.ToUpper(chromosomeClean), "MT") {
					chromosome = 0
				} else if strings.ToUpper(chromosomeClean) == "X" {
					chromosome = -1
				} else if strings.ToUpper(chromosomeClean) == "Y" {
					chromosome = -2
				} else {
					chromosome, _ = strconv.Atoi(chromosomeClean)
				}

				// clean start/end
				chromStartClean := strings.ReplaceAll(strings.ReplaceAll(_record[_startKey], ",", ""), " ", "")
				chromStart, _ = strconv.Atoi(chromStartClean)

				chromEndClean := strings.ReplaceAll(strings.ReplaceAll(_record[_endKey], ",", ""), " ", "")
				chromEnd, _ = strconv.Atoi(chromEndClean)

				discoveredGene := &models.Gene{
					Nomenclature: models.Nomenclature{
						Names:     names,
						GeneNames: geneNames,
					},
					Chrom:      chromosome,
					Start:      chromStart,
					End:        chromEnd,
					AssemblyId: _assId,
				}

				fmt.Printf("Keys :%d, %d, %d, %d, %d -- %s\n", _chromHeaderKey, _startKey, _endKey, _nameHeaderKeys, _geneNameHeaderKeys, discoveredGene)

				iz.GeneIngestionBulkIndexingQueue <- &structs.GeneIngestionQueueStructure{
					Gene:      discoveredGene,
					WaitGroup: _gwg,
				}
			}(record, chromHeaderKey, startKey, endKey, nameHeaderKeys, geneNameHeaderKeys, assId, &geneWg)

			// fmt.Printf("Stats : %d\n", iz.GeneIngestionBulkIndexer.Stats())
		}
		geneWg.Wait()

	}
}
