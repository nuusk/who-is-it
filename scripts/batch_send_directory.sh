#!/usr/bin/env bash

API_ENDPOINT=$1
DIRECTORY=$2

for file in "$DIRECTORY"/*
do
    file_name=$(basename $file)
    extension="${file_name##*.}"
    encoded=$(base64 -i $file)
    JSON="{\"imageBase64\": \"$encoded\", \"fileName\": \"$file_name\", \"extension\": \"$extension\"}"

    echo "Uploading file $file_name to $API_ENDPOINT"
    curl -XPOST -d "$JSON" -H "Content-type: application/json" -v $API_ENDPOINT
done

