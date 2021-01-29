using System;
using FolkerKinzel.VCards;
using Nest;

namespace Bento.Variants.Console
{
    class Program
    {
        static void Main(string[] args)
        {
            System.Console.WriteLine("Hello World!");

            // Establish connection with local Elasticsearch
            var url = "http://0.0.0.0:9200";
            var indexMap = "bento-variants-data";

            var settings = new ConnectionSettings(new Uri(url))
                .DefaultIndex(indexMap);

            var client = new ElasticClient(settings);

            // TODO: load 1000Genomes and push to Elasticsearch
            
        }
    }
}
