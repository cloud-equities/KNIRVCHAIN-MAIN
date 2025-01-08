module KNIRVCHAIN-MAIN

go 1.21

replace github.com/cloud-equities/KNIRVCHAIN-chain v1.2.3 => ./

require github.com/syndtr/goleveldb v1.0.0

require (
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/joho/godotenv v1.5.1
)
