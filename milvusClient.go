package main

import (
	"context"
	"os"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

// MilvusCli is the global shared milvus client instance
var MilvusCli client.Client

/**
* @description: Initialize the milvus client
* @param ctx context.Context
* @return client instance
 */
func InitMilvusClient(ctx context.Context) client.Client {
	newClient, err := client.NewClient(ctx, client.Config{
		Address: os.Getenv("MILVUS_ADDRESS"),
		DBName:  os.Getenv("MILVUS_DBNAME"),
	})
	if err != nil {
		panic(err)
	}
	MilvusCli = newClient
	return newClient
}

/**
* @description: Get the global shared milvus client instance
* @return the global shared milvus client instance
 */
func GetMilvusClient() client.Client {
	return MilvusCli
}
