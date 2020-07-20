package main

import (
	"ceylon/cli/mgt"
	"context"
)

func main() {

	deployManager := mgt.DeployManager{
		Context: context.Background(),
	}

	err := deployManager.Deploy()
	if err != nil {
		panic(err)
	}
}
