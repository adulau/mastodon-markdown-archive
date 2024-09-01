#!/bin/bash

while true; do
     command="./mastodon-markdown-archive --template=../files/templates/post.tmpl --user=https://paperbay.org/@a --dist=./posts --persist-last=./last  --max-id=$(test -f ./last && cat ./last || echo "") --download-media=./posts --porcelain=true --threaded=true"
     output=$($command)
     if [[ "$output" -eq 0 ]]; then
	    echo "No posts returned. Exiting"
	    break
     fi
     echo "Fetched $output posts. Continuing."
     sleep 1
done

