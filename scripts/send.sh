#!/usr/bin/env bash

API_ENDPOINT=$1
FILE=$2

FILENAME=$(basename $FILE)
echo "Uploading file $FILENAME to $API_ENDPOINT"

ENCODED=$(base64 -i $FILE)

JSON="{\"imageBase64\": \"$ENCODED\", \"fileName\": \"$FILENAME\"}"

curl -XPOST -d "$JSON" -H "Content-type: application/json" -v $API_ENDPOINT