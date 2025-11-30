#!/bin/sh
sleep 15
./migrate-app -command up
exec ./product-app
