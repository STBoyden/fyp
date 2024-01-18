module github.com/STBoyden/fyp/src/server

go 1.21.0

replace github.com/STBoyden/fyp/src/common/utils => ../common/utils

replace github.com/SToyden/fyp/src/common/game-state => ../common/game-state

require (
	github.com/STBoyden/fyp/src/common/utils v0.0.0-00010101000000-000000000000
	github.com/phakornkiong/go-pattern-match v1.1.0
)

require github.com/TwiN/go-color v1.4.1 // indirect
