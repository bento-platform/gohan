workflow vcf_gz {
    String gohan_url
    Array[File] vcf_gz_file_names
    String assembly_id
    String project_dataset
    String filter_out_references
    String access_token

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
    String gohan_url
    String vcf_gz_file_name
    String assembly_id
    String project
    String dataset
    String filter_out_references
    String access_token

    command {
        echo "Using temporary-token : ${access_token}"

        QUERY="fileNames=${vcf_gz_file_name}&assemblyId=${assembly_id}&dataset=${dataset}&project=${project}&filterOutReferences=${filter_out_references}"
        AUTH_HEADER="Authorization: Bearer ${access_token}"
        
        # TODO: refactor
        # append temporary-token header if present
        if [ "${access_token}" == "" ]
        then
            RUN_RESPONSE=$(curl -vvv "${gohan_url}/private/variants/ingestion/run?$QUERY" -k | sed 's/"/\"/g')
        else
            RUN_RESPONSE=$(curl -vvv -H "$AUTH_HEADER" "${gohan_url}/private/variants/ingestion/run?$QUERY" -k | sed 's/"/\"/g')
        fi
        
        echo $RUN_RESPONSE 

        # reformat response string to include double quotes in the json object
        RUN_RESPONSE_WITH_QUOTES=$(echo $RUN_RESPONSE | sed 's/"/\"/g')
        echo $RUN_RESPONSE_WITH_QUOTES

        # obtain request id from the response for this one file just requested to process
        REQUEST_ID=$(echo $RUN_RESPONSE_WITH_QUOTES | jq -r '.[] |"\(.id)"')
        echo $REQUEST_ID

        # give it a second..
        sleep 1s

        # "while loop to ping '/variants/ingestion/requests' and wait for this file ingestion to complete or display an error..."
        while :
        do
            # TODO: refactor
            # fetch run requests
            # append temporary-token header if present
            if [ "${access_token}" == "" ]
            then
                REQUESTS=$(curl -vvv "${gohan_url}/private/variants/ingestion/requests" -k)
            else
                REQUESTS=$(curl -vvv -H "$AUTH_HEADER" "${gohan_url}/private/variants/ingestion/requests" -k)
            fi

            echo $REQUESTS
            
            # reformat response string to include double quotes in the json object
            REQ_WITH_QUOTES=$(echo $REQUESTS | sed 's/"/\"/g')
            echo $REQ_WITH_QUOTES

            # organize json objects as individual lines per response object (file being processed)
            JQ_RES=$(echo $REQ_WITH_QUOTES | jq -r  '.[] | "\(.id) \(.filename) \(.state)"')
            echo "$JQ_RES"


            # determine the state of the run request by filename
            THIS_FILE_RESULT=$(echo "$JQ_RES" | grep $REQUEST_ID | tr ' ' '\n' | grep . | tail -n1)
            echo $THIS_FILE_RESULT
            
            if [ "$THIS_FILE_RESULT" == "Done" ] || [ "$THIS_FILE_RESULT" == "Error" ]
            then
                WITH_ERROR_MESSAGE=

                if [ "$THIS_FILE_RESULT" == "Error" ]
                then
                    WITH_ERROR_MESSAGE=" in error!" 
                    echo "This is what we found from the /variants/ingestion/requests :"
                    echo "$THIS_FILE_RESULT"
                fi

                echo "File ${vcf_gz_file_name} with assembly id ${assembly_id} done processing $WITH_ERROR_MESSAGE" 

                break
            elif [ "$THIS_FILE_RESULT" == "" ]
            then
                echo "Something went wrong. Got invalid response from Gohan API : $REQUESTS" 
                break
            else
                echo "Waiting 5 seconds.."
                sleep 5s
            fi
        done
    }
}
