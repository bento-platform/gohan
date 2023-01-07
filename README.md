# Gohan - A Genomic Variants API

<div style="text-align:center">
  <img src="https://www.publicdomainpictures.net/pictures/110000/velka/bowl-of-rice.jpg" alt="bowl-of-rice" width="50%" style="align:middle;"/>
</div>




## Prerequisites
- Golang >= 1.19
  - installation: https://golang.org/doc/install
  - other references 
    - https://linguinecode.com/post/install-golang-linux-terminal
- Elasticsearch
  - getting started: https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html
  - overview tutorial: https://www.youtube.com/watch?v=C3tlMqaNSaI
- Bash
  - overview: https://en.wikipedia.org/wiki/Bash_%28Unix_shell%29
- Make
  - overview: https://en.wikipedia.org/wiki/Make_(software)
- Docker
  - getting started: https://www.docker.com/get-started
  - other references:
    - linux installation: https://www.tutorialspoint.com/docker/installing_docker_on_linux.htm
    - linux running without sudo : https://dev.to/nabbisen/docker-without-sudo-34ci
- Docker-Compose
  - getting started: https://docs.docker.com/compose/gettingstarted/
- Visual Studio Code (recommended)
  - getting started: https://code.visualstudio.com/docs
  - Golang extension: https://marketplace.visualstudio.com/items?itemName=golang.go
- PERL (optional)
  - installation: https://learn.perl.org/installing/unix_linux.html
- htspwd
  - linux installation: https://www.digitalocean.com/community/tutorials/how-to-set-up-password-authentication-with-apache-on-ubuntu-18-04-quickstart
<br />


## TL;DR

### Typical use-case walkthrough
```
  # environment
  cp ./etc/example.env .env # modify to your needs

  # kickstart dockerized gohan environment
  make init

  # (optional): if you plan on modifying the api codebase before deploying
  make init-dev

  # gateway & certificates
  mkdir -p gateway/certs/dev

  openssl req -newkey rsa:2048 -nodes -keyout gateway/certs/dev/gohan_privkey1.key -x509 -days 365 -out gateway/certs/dev/gohan_fullchain1.crt
  openssl req -newkey rsa:2048 -nodes -keyout gateway/certs/dev/es_gohan_privkey1.key -x509 -days 365 -out gateway/certs/dev/es_gohan_fullchain1.crt


  # build services
  make build-gateway 
  make build-api 

  # run services
  make run-gateway
  make run-elasticsearch
  make run-drs
  make run-api
  
  
  # initiate genes catlogue:
  curl -k https://gohan.local/genes/ingestion/run
  
  # monitor progress:
  curl -k https://gohan.local/genes/ingestion/requests
  curl -k https://gohan.local/genes/ingestion/stats

  # view catalogue
  curl -k https://gohan.local/genes/overview


  # create table
  DATA='{
      "name": "Gohan Box Test Table",
      "data_type": "variant",
      "dataset": "00000000-0000-0000-0000-000000000000",
      "metadata": {}
  }'
  curl -k -0 -v -X POST https://gohan.local/tables \
    -H 'Content-Type:application/json' \
    --data "$(echo $DATA)" | jq

  # <obtain the table "id">


  # move vcf.gz files to `$GOHAN_API_VCF_PATH`

  # ingest vcf.gz
  curl -k https://gohan.local/variants/ingestion/run\?fileNames=<filename>\&assemblyId=GRCh37\&filterOutHomozygousReferences=true\&tableId=<table id>
  
  # monitor progress:
  curl -k https://gohan.local/variants/ingestion/requests
  curl -k https://gohan.local/variants/ingestion/stats

  # view variants
  curl -k https://gohan.local/variants/overview
```

## Getting started

### **Environment :**

First, from the project root, create a local file for environment variables with default settings by running

```
cp ./etc/example.env .env
```
 and make any necessary changes, such as the Elasticsearch `GOHAN_ES_USERNAME` and `GOHAN_ES_PASSWORD` when in production.
 > note: a known current bug is that GOHAN_ES_USERNAME must remain its default..

<br >

### **Initialization**
Run 
```
make init
```

<br />


### **Elasticsearch & Kibana :**

Run 
```
make run-elasticsearch 
```
and *(optionally)*
```
make run-kibana
```


<br />


### **DRS :**

Run 
```
make run-drs
```

<br />


### **Data Access Authorization with OPA (more on this to come..) :**

Run 
```
make build-authz
make run-authz
```

<br />
<br />


## Development

<div style="text-align:center">
  <img src="https://github.com/bento-platform/Gohan/blob/master/images/architecture.png?raw=true" alt="architecture" width="50%" style="align:middle;"/>
</div>

### **Gateway**
To create and use development certs from the project root, run
```
mkdir -p gateway/certs/dev

openssl req -newkey rsa:2048 -nodes -keyout gateway/certs/dev/gohan_privkey1.key -x509 -days 365 -out gateway/certs/dev/gohan_fullchain1.crt
openssl req -newkey rsa:2048 -nodes -keyout gateway/certs/dev/es_gohan_privkey1.key -x509 -days 365 -out gateway/certs/dev/es_gohan_fullchain1.crt
```

> Note: Ensure your `CN` matches the hostname (**gohan.local** by default)

These will be incorporated into the **Gateway** service (using NGINX by default, see `gateway/Dockerfile` and `gateway/nginx.conf` for details). Be sure to update your local `/etc/hosts` (on Linux) or `C:/System32/drivers/etc/hosts` (on Windows) file with the name of your choice.

Next, run
```
make build-gateway
make run-gateway
```

<br />


### **API**
<b>Containerized : </b>

&nbsp;To simply run a working instance of the api "out of the box", build the `docker image` and spawn the `container` with an fresh binary build by running

```
make build-api
make run-api
```

&nbsp;and the `docker-compose.yaml` file will handle the configuration.


<br />

<b>Local Development :</b>

&nbsp;This can be done multiple ways.

  1. `Terminal` : From the project root, run 
```
# load variables from local file
set -a
. ./.env
set +a

cd src/api

go run .
```

  2. `IDE (preferably VSCode)`

    - follow the recommended instructions listed at https://code.visualstudio.com/docs/languages/go

    - configure the `.vscode/launch.json` to inject the above mentioned variables as recommended by https://stackoverflow.com/questions/29971572/how-do-i-add-environment-variables-to-launch-json-in-vscode

    - click 'Run & Debug' > "Play" 

<b>Local Release</b>

&nbsp;To build / test from source; 

```
make build-api-local-binaries
```


&nbsp;The binary can then be found at *bin/api_${GOOS}_${GOARCH}* and executed locally with

```
# load variables from local file
set -a
. ./.env
set +a

# navigate to binary directory
cd bin/

# execute binary
./api_${GOOS}_${GOARCH}
```



<b>Endpoints :</b>

**`/variants`**


Request
> &nbsp;&nbsp;**GET** `/variants/overview`<br/>
> &nbsp;&nbsp;&nbsp;params: `none`

<br/>

Response
>```json
> {
>     "chromosomes": {
>         "<CHROMOSOME>": `number`,
>         ...
>     },
>     "sampleIDs": {
>         "<SAMPLEID>": `number`,
>         ...
>     },
>     "variantIDs": {
>         "<VARIANTID>": `number`,
>         ...
>     }
> }
>
>```
<br />

<b>Example :</b>
>```json
> {
>     "chromosomes": {
>         "21": 90548
>     },
>     "sampleIDs": {
>         "hg00096": 33664,
>         "hg00099": 31227,
>         "hg00111": 25657
>     },
>     "variantIDs": {
>         ".": 90548
>     }
> }
>
>```

<br />
<br />

Requests
> &nbsp;&nbsp;**GET** `/variants/get/by/variantId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **string** `( 1-23, X, Y, MT )`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele ( "A" | "C" | "G" | "T" | "N" or some combination thereof )`
>   - alternative : **string** `an allele`
>   - ids : **string** `(a comma-deliminated list of variant ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`
>   - includeInfoInResultSet : **boolean** `(true | false)`
>   - genotype : **string** `( "HETEROZYGOUS" | "HOMOZYGOUS_REFERENCE" | "HOMOZYGOUS_ALTERNATE" )`
>   - getSampleIdsOnly : **bool**  *`(optional) -  default: false  `*
>
> &nbsp;&nbsp;**GET** `/variants/count/by/variantId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **string** `( 1-23, X, Y, MT )`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele`
>   - alternative : **string** `an allele`
>   - ids : **string** `(a comma-deliminated list of variant ID alphanumeric codes)`
>   - genotype : **string** `( "HETEROZYGOUS" | "HOMOZYGOUS_REFERENCE" | "HOMOZYGOUS_ALTERNATE" )`

> &nbsp;&nbsp;**GET** `/variants/get/by/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **string** `( 1-23, X, Y, MT )`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele`
>   - alternative : **string** `an allele`
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`
>   - includeInfoInResultSet : **boolean** `(true | false)`
>   - genotype : **string** `( "HETEROZYGOUS" | "HOMOZYGOUS_REFERENCE" | "HOMOZYGOUS_ALTERNATE" )`
>
> &nbsp;&nbsp;**GET** `/variants/count/by/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **string** `( 1-23, X, Y, MT )`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele`
>   - alternative : **string** `an allele`
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`
>   - genotype : **string** `( "HETEROZYGOUS" | "HOMOZYGOUS_REFERENCE" | "HOMOZYGOUS_ALTERNATE" )`
>

<br />

Generalized Response Body Structure

>```json  
>{
>     "status":  `number` (200 - 500),
>     "message": `string` ("Success" | "Error"),
>     "results": [
>         {
>             "query":  `string`,       // reflective of the type of id queried for, i.e 'variantId:abc123', or 'sampleId:HG0001
>             "assemblyId": `string` ("GRCh38" | "GRCh37" | "NCBI36" | "Other"),    // reflective of the assembly id queried for
>             "count":  `number`,   // this field is only present when performing a COUNT query
>             "start":  `number`,   // reflective of the provided lowerBound parameter, 0 if none
>             "end":  `number`,     // reflective of the provided upperBound parameter, 0 if none
>             "chromosome":  `string`,       // reflective of the chromosome queried for
>             "calls": [            // this field is only present when performing a GET query
>                 {
>                    "id": `string`, // variantId
>                    "chrom":  `string`,
>                    "pos": `number`,
>                    "ref": `[]string`,  // list of alleles
>                    "alt": `[]string`,  // list of alleles
>                    "info": [
>                        {
>                            "id": `string`,
>                            "value": `string`,
>                        },
>                        ...
>                    ],
>                    "format":`string`,
>                    "qual": `number`,
>                    "filter": `string`,
>                    "sampleId": `string`,
>                    "genotype_type": `string ( "HETEROZYGOUS" | "HOMOZYGOUS_REFERENCE" | "HOMOZYGOUS_ALTERNATE" )`,
>                    "assemblyId": `string` ("GRCh38" | "GRCh37" | "NCBI36" | "Other"),
>                 },
>                 ...
>             ]
>         },
>     ]
> }
> ```

<br />

<b>Examples :</b>

- http://localhost:5000/variants/get/by/variantId?lowerBound=25911206&upperBound=45911206&size=1000&sortByPosition=desc

- http://localhost:5000/variants/get/by/variantId?chromosome=22&ids=rs587678958,rs549011611,rs567408969

<br />

- http://localhost:5000/variants/get/by/sampleId?ids=HG00097&size=1000&sortByPosition=asc
  
<br />

- http://localhost:5000/variants/count/by/variantId?chromosome=8

- http://localhost:5000/variants/count/by/variantId?chromosome=22&lowerBound=10000000&upperBound=25000000


<br />

- http://localhost:5000/variants/get/by/sampleId?chromosome=2&id=NA12815&size=10000

- http://localhost:5000/variants/get/by/sampleId?chromosome=2&id=NA12815&size=1000&lowerBound=1000&upperBound=100000


<br />


Request
> &nbsp;&nbsp;**GET** `/variants/ingestion/run`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - filename : **string** `(required)`

<br/>

Response
>```json  
> {
>     "state":  `number` ("Queuing" | "Running" | "Done" | "Error"),
>     "id": `string`,
>     "filename": `string`,
>     "message": `string`,
> }
> ```

<br />
<br />

Request
> &nbsp;&nbsp;**GET** `/variants/ingestion/requests`<br/>
> &nbsp;&nbsp;&nbsp;params: `none`

<br/>

Response
>```json  
> [
>   {
>     "state":  `number` ("Queuing" | "Running" | "Done" | "Error"),
>     "id": `string`,
>     "filename": `string`,
>     "message": `string`,
>     "createdAt": `timestamp string`,
>     "updatedAt": `timestamp string`
>   },
>   ...
> ]
> ```


<br />
<br />


**`/tables`**

<br />


Request
> &nbsp;&nbsp;**GET** `/tables`<br/>

<br/>

Response
>```json  
> [
>   {
> 	  "id":             `string`,
>  	  "name":           `string`,
> 	  "data_type":      `string`,
> 	  "dataset":        `string`,
> 	  "assembly_ids": `[]string`,
> 	  "metadata":        {...},
> 	  "schema":          {...},
>   },
>   ...
> ]
> ```

<br />
<br />


Request
> &nbsp;&nbsp;**POST** `/tables`<br/>
>```json
> {
>    "name":           `string`,
>    "data_type":      `string`,
>    "dataset":        `string`,
>    "metadata":        {...},
> }
> ```

<br/>

Response
>```json
> {
>    "id":             `string`,
>    "name":           `string`,
>    "data_type":      `string`,
>    "dataset":        `string`,
>    "assembly_ids": `[]string`,
>    "metadata":        {...},
>    "schema":          {...},
> }
> ```


<br />
<br />


Request
> &nbsp;&nbsp;**GET** `/tables/:id`<br/>
> &nbsp;&nbsp;&nbsp;path params: 
>   - id : **string (UUID)** `(required)`

<br/>

Response
>```json
> {
>    "id":             `string`,
>    "name":           `string`,
>    "data_type":      `string`,
>    "dataset":        `string`,
>    "assembly_ids": `[]string`,
>    "metadata":        {...},
>    "schema":          {...},
> }
> ```

<br />
<br />


Request
> &nbsp;&nbsp;**GET** `/tables/:id/summary`<br/>
> &nbsp;&nbsp;&nbsp;path params: 
>   - id : **string (UUID)** `(required)`

<br/>

Response
>```json
> {
>    "count":               `int`,
>    "data_type_specific":  {...},
> }
> ```

<br />
<br />


Request
> &nbsp;&nbsp;**DELETE** `/tables/:id`<br/>
> &nbsp;&nbsp;&nbsp;path params: 
>   - id : **string (UUID)** `(required)`

<br/>

Response

`Status Code:` **204**

<br />
<br />




## Deployments :

All in all, run 
```
make run-elasticsearch 
make run-drs
make build-gateway && make run-gateway 
make build-api && make run-api

# and optionally
make run-kibana
```
<br />

For other handy tools, see the Makefile. Among those already mentionned here, you'll find other `build`, `run`, `stop` and `clean-up` commands.


<br />

## Tests :

Once `elasticsearch`, `drs`, the `api`, and the `gateway` are up, run 
```
make test-api-dev
```

