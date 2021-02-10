## Prerequisites:

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

Run `docker-compose up -d`

<br />



### **Console :**

Purpose: to ingest a VCF into Elasticsearch.<br />
Copy the VCF to the Bento.Variants.Console directory, and run `dotnet run --project Bento.Variants.Console`.<br />
> Note: It is assumed Elasticsearch is running on localhost:9200

<br />



### **API :**

Run `dotnet run --project Bento.Variants.Api`

<b>Endpoints :</b>

&nbsp;&nbsp;**GET** /variants/get<br/>
&nbsp;&nbsp;&nbsp;params: 
  - chromosome : **number** `(default is "*" if not specified)`
  - lowerBound : **number**
  - upperBound : **number**
  - labels : **string** `(comma-deliminated list of variant alphanumeric codes)`
  - size : **number** `(maximum number of results per label if one or more labels are specified)`

> Note: the `lower/upperBound` and `labels` parameters used together is redundant and may result in clashing elasticsearch query logic

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
 
> Note: the `lower/upperBound` and `labels` parameters used together is redundant and may result in clashing elasticsearch query logic


<br />

<b>Examples :</b>

- http://localhost:5000/variants/get?lowerBound=25911206&upperBound=45911206&rowCount=1000

- http://localhost:5000/variants/get?chromosome=22&labels=rs587678958,rs549011611,rs567408969

<br />

- http://localhost:5000/variants/get/by/sampleIds?sampleIds=HG00097&rowCount=1000
  
<br />

- http://localhost:5000/variants/count?chromosome=8

- http://localhost:5000/variants/count?chromosome=22&lowerBound=10000000&upperBound=25000000