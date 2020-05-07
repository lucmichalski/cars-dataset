#!/bin/bash
url="http://www.cityfeet.com/cont/api/search/listings-spatial"
cookie="ASP.NET_SessionId=x335iekckm5tqxcq12psv1p2; __RequestVerificationToken_L2NvbnQ1=FTTyjLMPpvjTLNYvWo5a5yFqhos830-fpyjtxwr4vsVnG8P7_bf5zEEpH4JjY2KfIKgHMuuotd9IyW4iUmSeYRHnLzQ1"
DATE=`date +%Y-%m-%d`
#for i in $(cat query.txt); do
for i in {1..9}; do
    body="{'location':{'name':'San Francisco, CA','bb':[37.708131,-122.51777,37.863424,-122.3570311],'lat':37.7857775,'lng':-122.43740055,'state':'CA','city':'San Francisco','id':'3-19282','level':3},'lt':1,'pt':0,'sort':null,'partnerId':null,'lc':[],'mode':2,'portfolio':-1,'tt':0,'ignoreLocation':false,'KeyWord':null,'rent':{'type':1,'basis':0},'term':'San Francisco, CA','PageNum':$i,'PageSize':30,'state':{'\$type':'Cityfeet.Core.Listing.MultiSearchState, Core','ProviderPosition':{'PDS':$((30 * ($i -1))),'CF':0}}}"
    content="$(curl -v -s "$url" --header "Cookie: $cookie" --header "Content-Type: application/json" --data "$body" --cookie "$cookie")"
    echo "$content" > ./data/city-feet-com-listings-spatial-$DATE-$i.json
    sleep 5
done