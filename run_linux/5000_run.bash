#!/bin/bash
# go back
cd ../

# remove the file
rm -rf 5000

# Define the file path
file_path="constants/constants.go"

# Use sed to search and replace the value
sed -i 's/\(BLOCKCHAIN_DB_PATH\s*=\s*"\)[^\/]*\/knirvdb"/\15000\/knirvdb"/' "$file_path"

# run the file
go run main.go chain -port 5000 -miners_address knirvchain3dd025e8fec7eda7cdd012ddde9c8e978ee7fa33