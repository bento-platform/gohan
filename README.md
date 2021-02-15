## Prerequisites
- .NET Core 3.1
  - install: https://dotnet.microsoft.com/download/dotnet-core/3.1
- Elasticsearch
  - getting started: https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html
  - overview tutorial: https://www.youtube.com/watch?v=C3tlMqaNSaI
- Docker
  - getting started: https://www.docker.com/get-started

<br />
<br />


## Getting started

### **Environment :**

First, from the project root, create a local file for environment variables with default settings by running

> `cp ./etc/example.env .env`

 and make any necessary changes.

<br >


### **Elasticsearch & Kibana :**

Run 
> `make run-elasticsearch`

> `make run-kibana` *(optional)*

<br />
<br />


## Development

![Architecture](https://github.com/bento-platform/Bento.Variants/blob/master/images/architecture.png?raw=true)


### **Console**

*Purpose*: to ingest a set of VCFs into Elasticsearch.<br />
Copy the VCFs to a directory local to the project (*i.e. .../Bento.Variants/**vcfs***), and, from the project root, run 
> `dotnet run --project Bento.Variants.Console --vcfPath vcfs --elasticsearchURL http://localhost:9200`.

<br />


### **API**

From the project root, run 
> `dotnet run --project Bento.Variants.Api`

<b>Endpoints :</b>

> &nbsp;&nbsp;**GET** /variants/get/by/variantId<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(comma-deliminated list of variant ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`

<br/>

> &nbsp;&nbsp;**GET** /variants/get/by/sampleId<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`
>   - size : **number** `(maximum number of results per id)`
>   - sortByPosition : **string** `(<empty> | asc | desc)`

<br/>

> &nbsp;&nbsp;**GET** /variants/count/by/variantId<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(comma-deliminated list of variant ID alphanumeric codes)`

<br />


> &nbsp;&nbsp;**GET** /variants/count/by/sampleId<br/>
> &nbsp;&nbsp;&nbsp;params: 
>   - chromosome : **number** `(default is "*" if not specified)`
>   - lowerBound : **number**
>   - upperBound : **number**
>   - ids : **string** `(comma-deliminated list of sample ID alphanumeric codes)`

<br />

<b>Examples :</b>

- http://localhost:5000/variants/get/by/variantId?lowerBound=25911206&upperBound=45911206&size=1000&sortByPosition=desc

- http://localhost:5000/variants/get/by/variantId?chromosome=22&ids=rs587678958,rs549011611,rs567408969

<br />

- http://localhost:5000/variants/get/by/sampleId?ids=HG00097&size=1000&sortByPosition=asc
  
<br />

- http://localhost:5000/variants/count?chromosome=8

- http://localhost:5000/variants/count?chromosome=22&lowerBound=10000000&upperBound=25000000


<br />
<br />


## Releases
### **Console :**
Local Release: 

&nbsp;From ***Bento.Variants.Console/***, run 
> `dotnet publish -c Release --self-contained` 

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Bento.Variants.Console* and executed with

> `cd bin/Release/netcoreapp3.1/linux-x64/publish`
>
> `./Bento.Variants.Console --vcfPath vcfs --elasticsearchURL http://localhost:9200`

Local Alpine Release: 
> `dotnet publish -c ReleaseAlpine --self-contained` 

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Bento.Variants.Console*

> **Note:** this method is not recommended unless you are running your host machine on Alpine Linux. Unlike the **API** (seen below), this binary has no utility in being containerized. If you need to use this, run the same commands as you would with just a `Release` above but with `ReleaseAlpine` instead

<br />

### **API :**
Local Release: 

&nbsp;From ***Bento.Variants.Api/***, run 

> `dotnet publish -c Release --self-contained` 

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Bento.Variants.Api* and executed with

> `export ElasticSearch__PrimaryIndex=variants`<br />
> `export ElasticSearch__Protocol=http`<br />
> `export ElasticSearch__Host=localhost`<br />
> `export ElasticSearch__Port=9200`
>
> `cd bin/Release/netcoreapp3.1/linux-x64/publish`
>
> `./Bento.Variants.Api --urls http://localhost:5000`

<br />

Containerized Alpine Release: 

&nbsp; If all is well with the `Release`, from ***Bento.Variants.Api/***, run 

> `dotnet publish -c ReleaseAlpine --self-contained` 

&nbsp;The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Bento.Variants.Api*

&nbsp;When ready, build the `docker image` and spawn the `container` by running

> `make run-api`

&nbsp;and the `docker-compose.yaml` file will handle the configuration.

<br />
<br />



## Deployments :
### **Coming soon..**