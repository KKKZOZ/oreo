package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var operation = ""

func main() {
	flag.StringVar(&operation, "op", "", "operation")
	flag.Parse()

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(""),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: "http://localhost:8000"}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
	)
	if err != nil {
		panic(err)
	}
	client := dynamodb.NewFromConfig(cfg)
	tableName := "oreo"

	if operation == "reset" {
		fmt.Printf("Resetting table %s\n", tableName)
		if tableExists(client, tableName) {
			err = DeleteAllItems(client, tableName)
			if err != nil {
				log.Fatal("DeleteAllItems failed", err)
			}
		}
	}

	if operation == "delete" {
		fmt.Printf("Deleting table %s\n", tableName)
		_, err = client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			log.Fatal("DeleteTable failed", err)
		}
	}

	if operation == "create" {
		if tableExists(client, tableName) {
			fmt.Printf("Deleting table %s\n", tableName)
			_, err = client.DeleteTable(context.TODO(), &dynamodb.DeleteTableInput{
				TableName: aws.String(tableName),
			})
			if err != nil {
				log.Fatal("DeleteTable failed", err)
			}
		}

		fmt.Printf("Creating table %s\n", tableName)
		_, err = client.CreateTable(context.TODO(), &dynamodb.CreateTableInput{
			AttributeDefinitions: []types.AttributeDefinition{
				{
					AttributeName: aws.String("ID"),
					AttributeType: types.ScalarAttributeTypeS,
				},
			},
			KeySchema: []types.KeySchemaElement{
				{
					AttributeName: aws.String("ID"),
					KeyType:       types.KeyTypeHash,
				},
			},
			TableName: aws.String(tableName),
			// ProvisionedThroughput: &types.ProvisionedThroughput{
			// 	ReadCapacityUnits:  aws.Int64(10000),
			// 	WriteCapacityUnits: aws.Int64(10000),
			// },
			BillingMode: types.BillingModePayPerRequest, // 设置为按需模式
		})
		if err != nil {
			log.Fatal("CreateTable failed", err)
		}
	}
}
