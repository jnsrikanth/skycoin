package main

import (
	"github.com/skycoin/skycoin/src/cli"
	"github.com/skycoin/skycoin/src/skycoin"
)

func main() {
	/*
		skycoin.Run(&cli.DaemonArgs)
	*/

	/*
	   skycoin.Run(&cli.ClientArgs)

	   stop := make(chan int)
	   <-stop
	*/

	skycoin.Run(&cli.DevArgs)

}
