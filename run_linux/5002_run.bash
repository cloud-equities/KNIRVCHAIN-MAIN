#!/bin/bash
# go back
cd ../

# remove the file
#rm -rf 5002

# Define the file path
file_path="constants/constants.go"

# Use sed to search and replace the value
sed -i -E 's/\(BLOCKCHAIN_DB_PATH\s*=\s*"\)[^\/]*\/knirv.db"/\15002\/database/knirv.db"/' "$file_path"

# run the file
go run main.go chain -port 5002 -miners_address knirv.chain42d40be8b315e31dac50a4daf93ce366b1c62668 -remote_node http://127.0.0.1:5001