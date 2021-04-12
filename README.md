## Prerequisites
- .NET Core 3.1
  - installation: https://dotnet.microsoft.com/download/dotnet-core/3.1
- Elasticsearch
  - getting started: https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html
  - overview tutorial: https://www.youtube.com/watch?v=C3tlMqaNSaI
- Docker
  - getting started: https://www.docker.com/get-started
- Visual Studio Code (recommended)
  - getting started: https://code.visualstudio.com/docs
- PERL (optional)
  - installation: https://learn.perl.org/installing/unix_linux.html

<br />
<br />


## Getting started

### **Environment :**

First, from the project root, create a local file for environment variables with default settings by running

```
cp ./etc/example.env .env
```
 and make any necessary changes, such as the Elasticsearch `BENTO_VARIANTS_ES_USERNAME` and `BENTO_VARIANTS_ES_PASSWORD` when in production.

 > Note: if `BENTO_VARIANTS_ES_USERNAME` and `BENTO_VARIANTS_ES_PASSWORD` are to be modified for development, be sure to mirror the changes done in `.env` in the `Bento.Variants.Api/appsettings.Development.json` to give the API access, as the dev username and password is hard-coded in both files.

<br >


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
<br />


### **DRS :**

Run 
```
make run-drs
```

<br />
<br />


## Development

![Architecture](https://github.com/bento-platform/Bento.Variants/blob/master/images/architecture.png?raw=true)


### **Gateway**
To create and use development certs from the project root, run
```
mkdir -p gateway/certs/dev

openssl req -newkey rsa:2048 -nodes -keyout gateway/certs/dev/variants_privkey1.key -x509 -days 365 -out gateway/certs/dev/variants_fullchain1.crt
```

> Note: Ensure your `CN` matches the hostname (**variants.local** by default)

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

dotnet run --project Bento.Variants.Api
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

*Purpose*: to ingest a set of VCFs into Elasticsearch.

From the project root directory, copy your decompressed VCFs to a directory local to the console project (*i.e. ./Bento.Variants.Console/**vcfs***)

**(Recommended):** If you first want to split a compressed VCF `(*.vcf.gz)` that contains multiple samples into individual VCF files that only contain one sample each, move that file into the above mentionned directory local to the console project, and then from the project root, run


```
bash Bento.Variants.Console/preprocess.sh <ORIGINAL_VCF_GZ_FILEPATH>
```

> Note: preprocessing currently only works on **Linux** machines with **bash**

otherwise, just run 
```
source .env

dotnet clean
dotnet restore
dotnet build

dotnet run --project Bento.Variants.Console --vcfPath Bento.Variants.Console/vcfs \
  --elasticsearchUrl ${BENTO_VARIANTS_ES_PUBLIC_URL} \
  --elasticsearchUsername ${BENTO_VARIANTS_ES_USERNAME} \
  --elasticsearchPassword ${BENTO_VARIANTS_ES_PASSWORD} \
  --drsUrl ${BENTO_VARIANTS_DRS_PUBLIC_URL} \
  --documentBulkSizeLimit 100000

```
> Note: 
>
> on **Windows** machines, the vcfPath forward slashes above have to be converted to two backslashes, i.e.
>
>     Bento.Variants.Console\\vcfs
>
>
> `--documentBulkSizeLimit` is an optional flag! Tune it as you see fit to minimize ingestion time (`100000` is the default)

<br />
<br />

## Releases

### **API :**
Local Release: 

&nbsp;First, from ***Bento.Variants.Api/***, run 

```
dotnet clean
dotnet restore
```
```
dotnet publish -c Release --self-contained
```



&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Bento.Variants.Api* and executed with


```
export ElasticSearch__Username=${BENTO_VARIANTS_ES_USERNAME}
export ElasticSearch__Password=${BENTO_VARIANTS_ES_PASSWORD}
export ElasticSearch__GatewayPath=${BENTO_VARIANTS_ES_PUBLIC_GATEWAY_PATH}
export ElasticSearch__PrimaryIndex=${BENTO_VARIANTS_ES_PASSWORD}
export ElasticSearch__Protocol=${BENTO_VARIANTS_PUBLIC_PROTO}
export ElasticSearch__Host=${BENTO_VARIANTS_PUBLIC_HOSTNAME}
export ElasticSearch__Port=${BENTO_VARIANTS_PUBLIC_PORT}

cd bin/Release/netcoreapp3.1/linux-x64/publish

./Bento.Variants.Api --urls http://localhost:5000
```
<br />

Containerized Alpine Release: 

&nbsp; If all is well with the `Release`, from ***Bento.Variants.Api/***, run 

```
dotnet publish -c ReleaseAlpine --self-contained
```

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Bento.Variants.Api*

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

&nbsp;From ***Bento.Variants.Console/***, run 
```
dotnet clean
dotnet restore
```
```
dotnet publish -c Release --self-contained
```

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Bento.Variants.Console* and executed with

```
source ../.env
 
cd bin/Release/netcoreapp3.1/linux-x64/publish

./Bento.Variants.Console --vcfPath Bento.Variants.Console/vcfs \
  --elasticsearchUrl ${BENTO_VARIANTS_ES_PUBLIC_URL} \
  --elasticsearchUsername ${BENTO_VARIANTS_ES_USERNAME} \
  --elasticsearchPassword ${BENTO_VARIANTS_ES_PASSWORD} \
  --drsUrl ${BENTO_VARIANTS_DRS_PUBLIC_URL} 

```

Local Alpine Release: 
```
dotnet publish -c ReleaseAlpine --self-contained
```

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Bento.Variants.Console*

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
