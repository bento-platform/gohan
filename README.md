## Prerequisites
- .NET Core 3.1
  - installation: https://dotnet.microsoft.com/download/dotnet-core/3.1
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

![Architecture](https://github.com/bento-platform/Gohan/blob/master/images/architecture.png?raw=true)


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
dotnet clean
dotnet restore
dotnet build

dotnet run --project Gohan.Api
```

<b>Endpoints :</b>

***/variants*** <br />
Requests
> &nbsp;&nbsp;**GET** `/variants/get/by/variantId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(a comma-deliminated list of variant ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`
>   - includeSamplesInResultSet : **boolean** `(true | false)`
>
> &nbsp;&nbsp;**GET** `/variants/count/by/variantId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(a comma-deliminated list of variant ID alphanumeric codes)`

> &nbsp;&nbsp;**GET** `/variants/get/by/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`
>   - includeSamplesInResultSet : **boolean** `(true | false)`
>
> &nbsp;&nbsp;**GET** `/variants/count/by/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`
>
> &nbsp;&nbsp;**GET** `/variants/remove/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - id : **string** `(a single sample ID alphanumeric code)`

<br />

Generalized Response Structure

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
>                    "ref": `string`, ( "A" | "C" | "G" | "T" )
>                    "pos": `number`,
>                    "alt": `string`, ( "A" | "C" | "G" | "T" )
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


***/vcfs*** <br />
Request
> &nbsp;&nbsp;**GET** `/vcfs/get/by/sampleId`<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(required)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - id : **string** `(a single sample ID alphanumeric code)`
>   - size : **number** `(maximum number of results per id)`

<br/>

Response

*`- A VCF file -`*

<br />
<br />

### **Console**

*Purpose*: to ingest a set of VCFs into Elasticsearch and DRS.

From the project root directory, copy your compressed VCFs `(*.vcf.gz)` to a directory local to the console project (*i.e. ./Gohan.Console/**vcfs***)

**(Recommended):** If you first want to split a compressed VCF that contains multiple samples into individual VCF files that only contain one sample each, move that file into the above mentionned directory local to the console project, and then, from the project root, run


```
bash Gohan.Console/preprocess.sh Gohan.Console/vcfs/ORIGINAL.vcf.gz
```

> Note: preprocessing currently only works on **Linux** machines with **bash**

otherwise, just run 
```
source .env

dotnet clean
dotnet restore
dotnet build

dotnet run --project Gohan.Console --vcfPath Gohan.Console/vcfs \
  --elasticsearchUrl ${GOHAN_ES_PUBLIC_URL} \
  --elasticsearchUsername ${GOHAN_ES_USERNAME} \
  --elasticsearchPassword ${GOHAN_ES_PASSWORD} \
  --drsUrl ${GOHAN_DRS_PUBLIC_URL} \
  --drsUsername ${GOHAN_DRS_BASIC_AUTH_USERNAME} \
  --drsPassword ${GOHAN_DRS_BASIC_AUTH_PASSWORD} \
  --documentBulkSizeLimit 100000

```
> Note: 
>
> on **Windows** machines, the vcfPath forward slashes above have to be converted to two backslashes, i.e.
>
>     Gohan.Console\\vcfs
>
>
> `--documentBulkSizeLimit` is an optional flag! Tune it as you see fit to minimize ingestion time (`100000` is the default)

<br />
<br />

## Releases

### **API :**
Local Release: 

&nbsp;First, from ***Gohan.Api/***, run 

```
dotnet clean
dotnet restore
```
```
dotnet publish -c Release --self-contained
```



&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Gohan.Api* and executed with


```
export ElasticSearch__Username=${GOHAN_ES_USERNAME}
export ElasticSearch__Password=${GOHAN_ES_PASSWORD}
export ElasticSearch__GatewayPath=${GOHAN_ES_PUBLIC_GATEWAY_PATH}
export ElasticSearch__PrimaryIndex=${GOHAN_ES_PASSWORD}
export ElasticSearch__Protocol=${GOHAN_PUBLIC_PROTO}
export ElasticSearch__Host=${GOHAN_PUBLIC_HOSTNAME}
export ElasticSearch__Port=${GOHAN_PUBLIC_PORT}

cd bin/Release/netcoreapp3.1/linux-x64/publish

./Gohan.Api --urls http://localhost:5000
```
<br />

Containerized Alpine Release: 

&nbsp; If all is well with the `Release`, from ***Gohan.Api/***, run 

```
dotnet publish -c ReleaseAlpine --self-contained
```

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Gohan.Api*

&nbsp;When ready, build the `docker image` and spawn the `container` by running

```
make run-api
```
or
```
make run-api-alpine
```

&nbsp;and the `docker-compose.yaml` file will handle the configuration.

<br />

### **Console :**
Local Release: 

&nbsp;From ***Gohan.Console/***, run 
```
dotnet clean
dotnet restore
```
```
dotnet publish -c Release --self-contained
```

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Gohan.Console* and executed with

```
source ../.env
 
cd bin/Release/netcoreapp3.1/linux-x64/publish

./Gohan.Console --vcfPath Gohan.Console/vcfs \
  --elasticsearchUrl ${GOHAN_ES_PUBLIC_URL} \
  --elasticsearchUsername ${GOHAN_ES_USERNAME} \
  --elasticsearchPassword ${GOHAN_ES_PASSWORD} \
  --drsUrl ${GOHAN_DRS_PUBLIC_URL} \
  --drsUsername ${GOHAN_DRS_BASIC_AUTH_USERNAME} \
  --drsPassword ${GOHAN_DRS_BASIC_AUTH_PASSWORD} \
  --documentBulkSizeLimit 100000

```

Local Alpine Release: 
```
dotnet publish -c ReleaseAlpine --self-contained
```

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Gohan.Console*

> **Note:** this method is not recommended unless you are running your host machine on Alpine Linux. Unlike the **API** (seen below), this binary has no utility in being containerized. If you need to use this, run the same commands as you would with just a `Release` above but with `ReleaseAlpine` instead

<br />

<br />
<br />



## Deployments :

All in all, run 
```
make run-elasticsearch 
make build-gateway && make run-gateway 
make build-api && make run-api

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
