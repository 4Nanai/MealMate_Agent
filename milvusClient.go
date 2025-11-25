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
* @return nil if success, error if failed
 */
func InitMilvusClient(ctx context.Context) {
	newClient, err := client.NewClient(ctx, client.Config{
		Address: os.Getenv("MILVUS_ADDRESS"),
		DBName:  os.Getenv("MILVUS_DBNAME"),
	})
	if err != nil {
		panic(err)
	}
	MilvusCli = newClient
}

/**
* @description: Get the global shared milvus client instance
* @return the global shared milvus client instance
 */
func GetMilvusClient() client.Client {
	return MilvusCli
}
