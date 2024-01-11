module github.com/STBoyden/fyp/src/server

go 1.21.1

replace github.com/STBoyden/fyp/utils => ../utils

require github.com/STBoyden/fyp/src/utils v0.0.0-20240106154139-acf2a002cafc

require (
	github.com/TwiN/go-color v1.4.1 // indirect
	github.com/phakornkiong/go-pattern-match v1.1.0
)
