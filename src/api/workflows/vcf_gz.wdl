version 1.0

workflow vcf_gz {
    input {
        String gohan_url
        Array[File] vcf_gz_file_names
        String assembly_id
        String project_dataset
        Boolean filter_out_references
        String access_token
        Boolean validate_ssl
    }

    call project_and_dataset_id {
        input: project_dataset = project_dataset
    }

    scatter(file_name in vcf_gz_file_names) {
        call vcf_gz_gohan {
            input: gohan_url = gohan_url,
                   vcf_gz_file_name = file_name,
                   assembly_id = assembly_id,
                   project = project_and_dataset_id.out[0],
                   dataset = project_and_dataset_id.out[1],
                   filter_out_references = filter_out_references,
                   access_token = access_token,
                   validate_ssl = validate_ssl
        }
    }
}

task project_and_dataset_id {
    input {
        String project_dataset
    }
    command <<< python3 -c 'import json; print(json.dumps("~{project_dataset}".split(":")))' >>>
    output {
        Array[String] out = read_json(stdout())
    }
}

task vcf_gz_gohan {
    input {
        String gohan_url
        String vcf_gz_file_name
        String assembly_id
        String project
        String dataset
        Boolean filter_out_references
        String access_token
        Boolean validate_ssl
    }

    command <<<
        QUERY='fileNames=~{vcf_gz_file_name}&assemblyId=~{assembly_id}&dataset=~{dataset}&project=~{project}&filterOutReferences=~{true="true" false="false" filter_out_references}'

        AUTH_HEADER='Authorization: Bearer ~{access_token}'
        
        RUN_RESPONSE=$(curl -vvv \
            -H "${AUTH_HEADER}" \
            ~{true="" false="-k" validate_ssl} \
            "~{gohan_url}/private/variants/ingestion/run?${QUERY}" | sed 's/"/\"/g')
        
        echo "${RUN_RESPONSE}"

        # reformat response string to include double quotes in the json object
        RUN_RESPONSE_WITH_QUOTES=$(echo $RUN_RESPONSE | sed 's/"/\"/g')
        echo "${RUN_RESPONSE_WITH_QUOTES}"

        # obtain request id from the response for this one file just requested to process
        REQUEST_ID=$(echo $RUN_RESPONSE_WITH_QUOTES | jq -r '.[] |"\(.id)"')
        echo "${REQUEST_ID}"

        # give it a second..
        sleep 1s

        # "while loop to ping '/variants/ingestion/requests' and wait for this file ingestion to complete or display an error..."
        while :
        do
            # fetch run requests
            REQUESTS=$(curl -vvv \
                -H "${AUTH_HEADER}" \
                ~{true="" false="-k" validate_ssl} \
                "~{gohan_url}/private/variants/ingestion/requests")

            echo "${REQUESTS}"
            
            # reformat response string to include double quotes in the json object
            REQ_WITH_QUOTES=$(echo $REQUESTS | sed 's/"/\"/g')
            echo "${REQ_WITH_QUOTES}"

            # organize json objects as individual lines per response object (file being processed)
            JQ_RES=$(echo $REQ_WITH_QUOTES | jq -r  '.[] | "\(.id) \(.filename) \(.state)"')
            echo "${JQ_RES}"

            # determine the state of the run request by filename
            THIS_FILE_RESULT=$(echo "$JQ_RES" | grep $REQUEST_ID | tr ' ' '\n' | grep . | tail -n1)
            echo "${THIS_FILE_RESULT}"
            
            if [ "${THIS_FILE_RESULT}" == "Done" ] || [ "${THIS_FILE_RESULT}" == "Error" ]; then
                WITH_ERROR_MESSAGE=''

                if [ "${THIS_FILE_RESULT}" == "Error" ]; then
                    WITH_ERROR_MESSAGE=" in error!" 
                    echo "This is what we found from the /variants/ingestion/requests :"
                    echo "${THIS_FILE_RESULT}"
                fi

                echo "File ~{vcf_gz_file_name} with assembly id ~{assembly_id} done processing ${WITH_ERROR_MESSAGE}"

                break
            elif [ "${THIS_FILE_RESULT}" == "" ]; then
                echo "Something went wrong. Got invalid response from Gohan API: ${REQUESTS}"
                break
            else
                echo '~{vcf_gz_file_name}: Waiting 5 seconds...'
                sleep 5s
            fi
        done
    >>>

    output {
        String out = stdout()
        String err = stderr()
    }
}
