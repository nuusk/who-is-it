#!/usr/bin/env bash

api=$1
directory=$2

function post {
    file_name=$(basename $1)
    extension="${file_name##*.}"
    echo "Uploading file $file_name to $api"
    (echo -n '{"imageBase64": "'; base64 $file; echo '", "fileName":"'$file_name'", "extension":"'$extension'"}') | curl -H "Content-Type: application/json" -d @- $api
}  

for file in "$directory"/*; do
    post $file &
done
