workflow vcf_gz {
    String gohan_url
    Array[String] vcf_gz_file_names
    String assembly_id
    String drs_file_id

    scatter(file_name in vcf_gz_file_names) {
        call vcf_gz_gohan {
            input: gohan_url = gohan_url,
                   vcf_gz_file_name = file_name,
                   assembly_id = assembly_id,
                   drs_file_id = drs_file_id
        }
    }
}

task vcf_gz_gohan {
    String gohan_url
    String vcf_gz_file_name
    String assembly_id
    String drs_file_id
    command {
        curl "${gohan_url}/variants/ingest/run?fileName=${vcf_gz_file_name}&assemblyId=${assembly_id}drsFileId=${drs_file_id}"
        # "while loop to ping '/variants/ingest/requests' and waiting for this file ingestion to complete or display an error..."
    }
    output {}
}
