package main

import (
	"fmt"
	"apigw-contrib/util/conf"
	"./pb"
)

type OneOperationInfo struct {
	OpType	string
}

type PresscallInfo struct {
	OpList []OneOperationInfo
}

var PressConf PresscallInfo

func main() {
	//init conf
	err := conf.InitConfig("apigw.conf", PressConf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	tgwadm.CtrlMsg ctrlMsg

	
	fmt.Println("Hello world.")
}