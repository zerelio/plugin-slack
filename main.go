package main

import sdk "github.com/zerelio/plugin-sdk"

func main() {
	sdk.Serve(&SlackPlugin{})
}
