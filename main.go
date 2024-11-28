package main

import (
	"fmt"
	"image-syncer/cmd"
	"image-syncer/pkg/client"
	"image-syncer/pkg/utils"
	"strings"
)

func main1() {
	cmd.Execute()
}

// const HubAuthFile = "~/.docker/config.json"
const HubAuthFile = "config.json"

func DoOne(src, dst string) error {
	var imagelist map[string]string
	imagelist = make(map[string]string)
	imagelist[src] = dst

	for k, v := range imagelist {
		fmt.Printf("src: %s, dst: %s\n", k, v)
	}
	// work starts here
	client, err := client.NewSyncClient("", HubAuthFile, imagelist, "./", "", "yaml",
		1, 1, utils.RemoveEmptyItems([]string{}), utils.RemoveEmptyItems([]string{}), false)
	if err != nil {
		return fmt.Errorf("init sync client error: %v", err)
	}

	return client.Run()
}

func GetDestName(src string) string {
	dstPreFix := "registry.cn-hangzhou.aliyuncs.com/benz/ff"
	//dstPreFix := "hub.hitry.io/ff/bbb"
	dstHou := src[strings.LastIndex(src, "/")+1:]
	dstHou = strings.Replace(dstHou, ":", "_", -1)
	dstName := dstPreFix + ":" + dstHou
	return dstName
}

func main() {
	//DoOne("hub.hitry.io/ff/aaa:v1", "hub.hitry.io/ff/bbb:v2")
	//DoOne("hub.hitry.io/hitry/auth:h1.1.327332", "hub.hitry.io/ff/bbb:v2")

	//src := "hub.hitry.io/hitry/auth:h1.1.327332"

	// DoOne(src, GetDestName(src))
	// fmt.Println("src:", src)
	// fmt.Println("dst:", GetDestName(src))
	hs := &HttpServer{}
	hs.Run()
}
