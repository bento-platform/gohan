# Gohan - A Genomic Variants API

<div style="text-align:center">
  <img src="https://www.publicdomainpictures.net/pictures/110000/velka/bowl-of-rice.jpg" alt="bowl-of-rice" width="50%" style="align:middle;"/>
</div>




## Prerequisites
- Golang >= 1.14.2
  - installation: https://golang.org/doc/install
- UPX
  - docs: https://github.com/upx/upx
- Elasticsearch
  - getting started: https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html
  - overview tutorial: https://www.youtube.com/watch?v=C3tlMqaNSaI
- Bash
  - overview: https://en.wikipedia.org/wiki/Bash_%28Unix_shell%29
- Make
  - overview: https://en.wikipedia.org/wiki/Make_(software)
- Docker
  - getting started: https://www.docker.com/get-started
- Docker-Compose
  - getting started: https://docs.docker.com/compose/gettingstarted/
- Visual Studio Code (recommended)
  - getting started: https://code.visualstudio.com/docs
- PERL (optional)
  - installation: https://learn.perl.org/installing/unix_linux.html

<br />


## Getting started

### **Environment :**

First, from the project root, create a local file for environment variables with default settings by running

```
cp ./etc/example.env .env
```
 and make any necessary changes, such as the Elasticsearch `GOHAN_ES_USERNAME` and `GOHAN_ES_PASSWORD` when in production.

 > Note: if `GOHAN_ES_USERNAME` and `GOHAN_ES_PASSWORD` are to be modified for development, be sure to mirror the changes done in `.env` in the `Gohan.Api/appsettings.Development.json` to give the API access, as the dev username and password is hard-coded in both files.

<br >

### **Init**
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

The first startup may fail on an `AccessDeniedException[/usr/share/elasticsearch/data/nodes];` and can be resolved by setting the data directory to have less strict permissions with
```
sudo chmod -R 777 data/
```

<br />


### **DRS :**

Run 
```
make build-drs
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
```

> Note: Ensure your `CN` matches the hostname (**gohan.local** by default)

These will be incorporated into the **Gateway** service (using NGINX by default, see `gateway/Dockerfile` and `gateway/nginx.conf` for details). Be sure to update your local `/etc/hosts` (on Linux) or `C:/System32/drivers/etc/hosts` (on Windows) file with the name of your choice.

Next, run
```
make build-gateway
make run-gateway
```

<br />
<br />


### **API**

From the project root, run 
```
export GOHAN_API_INTERNAL_PORT=${GOHAN_API_INTERNAL_PORT}
export GOHAN_API_VCF_PATH=${GOHAN_API_VCF_PATH}

# Elasticsearch
export GOHAN_ES_URL=${GOHAN_PRIVATE_ES_URL}
export GOHAN_ES_USERNAME=${GOHAN_ES_USERNAME}
export GOHAN_ES_PASSWORD=${GOHAN_ES_PASSWORD}

# AuthX
export GOHAN_AUTHZ_ENABLED=${GOHAN_API_AUTHZ_ENABLED}
export GOHAN_PUBLIC_AUTHN_JWKS_URL=${GOHAN_PUBLIC_AUTHN_JWKS_URL}
export GOHAN_PRIVATE_AUTHZ_URL=${GOHAN_PRIVATE_AUTHZ_URL}
export GOHAN_AUTHZ_REQHEADS=${GOHAN_API_AUTHZ_REQHEADS}

# DRS
export GOHAN_DRS_URL=${GOHAN_PRIVATE_DRS_URL}
export GOHAN_DRS_BASIC_AUTH_USERNAME=${GOHAN_DRS_BASIC_AUTH_USERNAME}
export GOHAN_DRS_BASIC_AUTH_PASSWORD=${GOHAN_DRS_BASIC_AUTH_PASSWORD}

cd src/api

go run .
```


<b>Endpoints :</b>

***/variants*** <br />
Requests
> &nbsp;&nbsp;**GET** `/variants/get/by/variantId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number**
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele ( "A" | "C" | "G" | "T"  or some combination thereof)`
>   - alternative : **string** `an allele`
>   - ids : **string** `(a comma-deliminated list of variant ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`
>   - includeSamplesInResultSet : **boolean** `(true | false)`
>
> &nbsp;&nbsp;**GET** `/variants/count/by/variantId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number**
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele`
>   - alternative : **string** `an allele`
>   - ids : **string** `(a comma-deliminated list of variant ID alphanumeric codes)`

> &nbsp;&nbsp;**GET** `/variants/get/by/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number**
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele`
>   - alternative : **string** `an allele`
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`
>   - includeSamplesInResultSet : **boolean** `(true | false)`
>
> &nbsp;&nbsp;**GET** `/variants/count/by/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number**
>   - lowerBound : **number**
>   - upperBound : **number**
>   - reference : **string** `an allele`
>   - alternative : **string** `an allele`
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`
>

<br />

Generalized Response Body Structure

>```json  
>{
>     "status":  `number` (200 - 500),
>     "message": `string` ("Success" | "Error"),
>     "data": [
>         {
>             "variantId":  `string`,
>             "sampleId":  `string`,
>             "count":  `number`,
>             "results": [
>                 {
>                    "filter": `string`,
>                    "pos": `number`,
>                    "ref": [
>                        `string`,  // an allele
>                    ],
>                    "alt": [
>                         `string`,  // an allele
>                    ],
>                    "info": [
>                        {
>                            "id": `string`,
>                            "value": `string`,
>                        },
>                        ...
>                    ],
>                    "format":`string`,
>                    "qual": `number`,
>                    "id": `string`,
>                    "samples": [
>                        {
>                            "sampleId": `string`,
>                            "variation": `string`,
>                        },
>                        ...
>                    ]
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

- http://localhost:5000/vcfs/get/by/sampleId?chromosome=2&id=NA12815&size=10000

- http://localhost:5000/vcfs/get/by/sampleId?chromosome=2&id=NA12815&size=1000&lowerBound=1000&upperBound=100000


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


## Releases

### **API :**
Local Release: 

&nbsp;From the project root, run 

```
make build-api-go-alpine-binaries
```



&nbsp;The binary can then be found at *bin/api_${GOOS}_${GOARCH}* and executed with

```
export GOHAN_API_INTERNAL_PORT=${GOHAN_API_INTERNAL_PORT}
export GOHAN_API_VCF_PATH=${GOHAN_API_VCF_PATH}

# Elasticsearch
export GOHAN_ES_URL=${GOHAN_PRIVATE_ES_URL}
export GOHAN_ES_USERNAME=${GOHAN_ES_USERNAME}
export GOHAN_ES_PASSWORD=${GOHAN_ES_PASSWORD}

# AuthX
export GOHAN_AUTHZ_ENABLED=${GOHAN_API_AUTHZ_ENABLED}
export GOHAN_PUBLIC_AUTHN_JWKS_URL=${GOHAN_PUBLIC_AUTHN_JWKS_URL}
export GOHAN_PRIVATE_AUTHZ_URL=${GOHAN_PRIVATE_AUTHZ_URL}
export GOHAN_AUTHZ_REQHEADS=${GOHAN_API_AUTHZ_REQHEADS}

# DRS
export GOHAN_DRS_URL=${GOHAN_PRIVATE_DRS_URL}
export GOHAN_DRS_BASIC_AUTH_USERNAME=${GOHAN_DRS_BASIC_AUTH_USERNAME}
export GOHAN_DRS_BASIC_AUTH_PASSWORD=${GOHAN_DRS_BASIC_AUTH_PASSWORD}

cd bin/

./api_${GOOS}_${GOARCH}
```

<br />

Containerized Alpine Release: 

&nbsp;When ready, build the `docker image` and spawn the `container` by running

```
make build-api-go-alpine-container
make run-api-go-alpine
```

&nbsp;and the `docker-compose.yaml` file will handle the configuration.

<br />
<br />



## Deployments :

All in all, run 
```
make run-elasticsearch 
make run-drs
make build-gateway && make run-gateway 
make build-api-go-alpine && make run-api-go-alpine

# and optionally
make run-kibana
```
<br />

For other handy tools, see the Makefile. Among those already mentionned here, you'll find other `build`, `run`, `stop` and `clean-up` commands.


<br />

## Tests :

Once `elasticsearch`, the `api`, and the `gateway` are up, run 
```
make test-api-dev
```
