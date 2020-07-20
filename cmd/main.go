package main

import "ceylon/cli/mgt"

func main() {

	deployManager := mgt.DeployManager{}

	err := deployManager.Deploy()
	if err != nil {
		panic(err)
	}
}
