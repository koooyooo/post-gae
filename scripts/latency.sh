#!/bin/sh
if [ $1 = "-l" ]; then
    curl -i http://localhost:8080/ -s -o /dev/null -w  "%{time_starttransfer}\n"
else
    curl -i https://dm-on-priv.appspot.com/ -s -o /dev/null -w  "%{time_starttransfer}\n"
fi