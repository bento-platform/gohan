## Prerequisites
- .NET Core 3.1
  - install: https://dotnet.microsoft.com/download/dotnet-core/3.1
- Elasticsearch
  - getting started: https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started.html
  - overview tutorial: https://www.youtube.com/watch?v=C3tlMqaNSaI
- Docker
  - getting started: https://www.docker.com/get-started

<br /><br />


## Getting started

### **Elasticsearch & Kibana :**

Run 
- `make run-dev-elasticsearch`
- `make run-dev-kibana` *(optional)*

<br />


## Development

### **Console**

*Purpose*: to ingest a set of VCFs into Elasticsearch.<br />
Copy the VCFs to a directory local to the project (*i.e. .../Bento.Variants/**vcfs***), and, from the project root, run 
- `dotnet run --project Bento.Variants.Console --vcfPath vcfs --elasticsearchURL http://localhost:9200`.

<br />


### **API**

From the project root, run 
- `dotnet run --project Bento.Variants.Api`

<b>Endpoints :</b>

&nbsp;&nbsp;**GET** /variants/get/by/variantIds<br/>
&nbsp;&nbsp;&nbsp;params: 
  - chromosome : **number** `(default is "*" if not specified)`
  - lowerBound : **number**
  - upperBound : **number**
  - variantIds : **string** `(comma-deliminated list of variant alphanumeric codes)`
  - size : **number** `(maximum number of results per label if one or more labels are specified)`

<br/>

&nbsp;&nbsp;**GET** /variants/get/by/sampleIds<br/>
&nbsp;&nbsp;&nbsp;params: 
  - sampleIds : **string** `(comma-deliminated list of sampleId alphanumeric codes)`
  - size : **number** `(maximum number of results per label if one or more labels are specified)`

<br/>

&nbsp;&nbsp;**GET** /variants/count<br/>
&nbsp;&nbsp;&nbsp;params: 
  - chromosome : **number** `(default is "*" if not specified)`
  - lowerBound : **number**
  - upperBound : **number**
  - labels : **string** `(comma-deliminated list of variant alphanumeric codes)`

<br />

<b>Examples :</b>

- http://localhost:5000/variants/get/by/variantIds?lowerBound=25911206&upperBound=45911206&rowCount=1000

- http://localhost:5000/variants/get/by/variantIds?chromosome=22&variantIds=rs587678958,rs549011611,rs567408969

<br />

- http://localhost:5000/variants/get/by/sampleIds?sampleIds=HG00097&rowCount=1000
  
<br />

- http://localhost:5000/variants/count?chromosome=8

- http://localhost:5000/variants/count?chromosome=22&lowerBound=10000000&upperBound=25000000


<br />


## Releases
### **Console :**
Local Release: 

&nbsp;From ***Bento.Variants.Console/***, run 
- `dotnet publish -c Release --self-contained` 
> The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Bento.Variants.Console*

&nbsp;Containerized Release: 
- `dotnet publish -c ReleaseAlpine --self-contained` 
> The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Bento.Variants.Console*

<br />

### **API :**
Local Release: 

&nbsp;From ***Bento.Variants.Api/***, run 
- `dotnet publish -c Release --self-contained` 
> The binary can then be found at *bin/Release/netcoreapp3.1/**linux-x64**/publish/Bento.Variants.Api*

&nbsp;Containerized Release: 
- `dotnet publish -c ReleaseAlpine --self-contained` 
> The binary can then be found at *bin/Release/netcoreapp3.1/**linux-musl-x64**/publish/Bento.Variants.Api*<br /><br />Once the release is ready, build the docker image and spawn the container by running
> - `make run-dev-api` 

<br />


## Deployments :
### **Coming soon..**