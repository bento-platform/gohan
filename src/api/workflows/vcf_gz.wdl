workflow vcf_gz {
    String gohan_url
    Array[String] vcf_gz_file_names
    String assembly_id

    scatter(file_name in vcf_gz_file_names) {
        call vcf_gz_gohan {
            input: gohan_url = gohan_url,
                   vcf_gz_file_name = file_name,
                   assembly_id = assembly_id,
        }
    }
}

task vcf_gz_gohan {
    String gohan_url
    String vcf_gz_file_name
    String assembly_id
    command {
        curl "${gohan_url}/variants/ingest/run?fileName=${vcf_gz_file_name}&assemblyId=${assembly_id}"

        # give it a second..
        sleep 1s

        # "while loop to ping '/variants/ingest/requests' and wait for this file ingestion to complete or display an error..."
        while
        do
            THIS_FILE_RESULT=$(curl "${gohan_url}/variants/ingestion/requests" -k | jq -r  '.[] | "\(.filename) \(.state)"' | grep v38 | awk '{print $2}')
            
            if [ $THIS_FILE_RESULT == "Done" || $THIS_FILE_RESULT == "Error" ]
                WITH_ERROR_MESSAGE=

                if [ $THIS_FILE_RESULT == "Error" ]
                    WITH_ERROR_MESSAGE=" in error!"
                    echo "This is what we found from the /variants/ingestion/requests :"
                    echo "${THIS_FILE_RESULT}"
                fi

                echo "File ${vcf_gz_file_name} with assembly id ${assembly_id} done processing${WITH_ERROR_MESSAGE}"

                break
            fi
        done
    }
    output {}
}
