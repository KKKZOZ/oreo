package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/ratelimit"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-errors/errors"
	"github.com/oreo-dtx-lab/oreo/internal/util"
	oreoconfig "github.com/oreo-dtx-lab/oreo/pkg/config"
	"github.com/oreo-dtx-lab/oreo/pkg/logger"
	"github.com/oreo-dtx-lab/oreo/pkg/txn"
)

var _ txn.Connector = (*DynamoDBConnection)(nil)

type KeyValueItem struct {
	ID    string `dynamodbav:"ID"`
	Value string `dynamodbav:"Value"`
}

type DynamoDBConnection struct {
	client       *dynamodb.Client
	tableName    string
	config       ConnectionOptions
	hasConnected bool
}

type ConnectionOptions struct {
	Region      string
	TableName   string
	Endpoint    string
	Credentials aws.CredentialsProvider
}

func NewDynamoDBConnection(config *ConnectionOptions) *DynamoDBConnection {
	if config == nil {
		config = &ConnectionOptions{
			Region:    "us-west-2",
			TableName: "oreo",
			Endpoint:  "http://localhost:8000",
		}
	}

	return &DynamoDBConnection{
		config:       *config,
		hasConnected: false,
	}
}

func (d *DynamoDBConnection) Connect() error {
	if d.hasConnected {
		return nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(""),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{URL: d.config.Endpoint}, nil
			})),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID: "dummy", SecretAccessKey: "dummy", SessionToken: "dummy",
				Source: "Hard-coded credentials; values are irrelevant for local DynamoDB",
			},
		}),
		config.WithRetryer(func() aws.Retryer {
			return retry.NewStandard(func(o *retry.StandardOptions) {
				o.RateLimiter = ratelimit.NewTokenRateLimit(10000)
			})
		}),
	)
	if err != nil {
		panic(err)
	}
	d.client = dynamodb.NewFromConfig(cfg)
	d.tableName = d.config.TableName
	d.hasConnected = true

	return nil
}

func (d *DynamoDBConnection) Close() error {
	d.hasConnected = false
	return nil
}

func (d *DynamoDBConnection) GetItem(key string) (txn.DataItem, error) {
	if !d.hasConnected {
		return &DynamoDBItem{}, errors.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	result, err := d.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return &DynamoDBItem{}, err
	}

	if result.Item == nil {
		return &DynamoDBItem{}, errors.New(txn.KeyNotFound)
	}

	var item DynamoDBItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return &DynamoDBItem{}, err
	}

	return &item, nil
}

func (d *DynamoDBConnection) PutItem(key string, value txn.DataItem) (string, error) {
	if !d.hasConnected {
		return "", errors.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	av, err := attributevalue.MarshalMap(value)
	if err != nil {
		logger.Log.Errorw("failed to marshal data item", "error", err)
		return "", err
	}

	_, err = d.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item:      av,
	})
	if err != nil {
		logger.Log.Errorw("failed to put item", "error", err)
		return "", err
	}

	return "", nil
}

func (d *DynamoDBConnection) ConditionalUpdate(
	key string,
	value txn.DataItem,
	doCreat bool,
) (string, error) {
	if !d.hasConnected {
		return "", errors.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	if doCreat {
		return d.atomicCreateDynamoItem(key, value)
	}

	newVer := util.AddToString(value.Version(), 1)

	updateExpr := "SET #val = :val, #gkl = :gkl, #ts = :ts, #tv = :tv, " +
		"#tl = :tl, #prev = :prev, #ll = :ll, #id = :id, #ver = :ver"

	exprAttrNames := map[string]string{
		"#val":  "Value",
		"#gkl":  "GroupKeyList",
		"#ts":   "TxnState",
		"#tv":   "TValid",
		"#tl":   "TLease",
		"#prev": "Prev",
		"#ll":   "LinkedLen",
		"#id":   "IsDeleted",
		"#ver":  "Version",
	}

	exprAttrValues := map[string]types.AttributeValue{
		":val":    &types.AttributeValueMemberS{Value: value.Value()},
		":gkl":    &types.AttributeValueMemberS{Value: value.GroupKeyList()},
		":ts":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", value.TxnState())},
		":tv":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", value.TValid())},
		":tl":     &types.AttributeValueMemberS{Value: value.TLease().Format(time.RFC3339Nano)},
		":prev":   &types.AttributeValueMemberS{Value: value.Prev()},
		":ll":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", value.LinkedLen())},
		":id":     &types.AttributeValueMemberBOOL{Value: value.IsDeleted()},
		":ver":    &types.AttributeValueMemberS{Value: newVer},
		":oldver": &types.AttributeValueMemberS{Value: value.Version()},
	}

	_, err := d.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: key},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
		ConditionExpression:       aws.String("#ver = :oldver"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}

	return newVer, nil
}

func (d *DynamoDBConnection) ConditionalCommit(
	key string,
	version string,
	tCommit int64,
) (string, error) {
	if !d.hasConnected {
		return "", errors.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	newVer := util.AddToString(version, 1)

	updateExpr := "SET #ts = :ts, #tv = :tv, #ver = :ver"
	exprAttrNames := map[string]string{
		"#ts":  "TxnState",
		"#tv":  "TValid",
		"#ver": "Version",
	}

	exprAttrValues := map[string]types.AttributeValue{
		":ts":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", oreoconfig.COMMITTED)},
		":tv":     &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", tCommit)},
		":ver":    &types.AttributeValueMemberS{Value: newVer},
		":oldver": &types.AttributeValueMemberS{Value: version},
	}

	_, err := d.client.UpdateItem(context.Background(), &dynamodb.UpdateItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: key},
		},
		UpdateExpression:          aws.String(updateExpr),
		ExpressionAttributeNames:  exprAttrNames,
		ExpressionAttributeValues: exprAttrValues,
		ConditionExpression:       aws.String("#ver = :oldver"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}

	return newVer, nil
}

func (d *DynamoDBConnection) AtomicCreate(key string, value any) (string, error) {
	if !d.hasConnected {
		return "", errors.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	str := util.ToString(value)
	_, err := d.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item: map[string]types.AttributeValue{
			"ID":    &types.AttributeValueMemberS{Value: key},
			"Value": &types.AttributeValueMemberS{Value: str},
		},
		ConditionExpression: aws.String("attribute_not_exists(ID)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			// Key exists, get the current value
			result, err := d.Get(key)
			if err != nil {
				return "", err
			}
			return result, errors.New(txn.KeyExists)
		}
		return "", err
	}

	return "", nil
}

func (d *DynamoDBConnection) atomicCreateDynamoItem(
	key string,
	value txn.DataItem,
) (string, error) {
	newVer := util.AddToString(value.Version(), 1)

	av, err := attributevalue.MarshalMap(DynamoDBItem{
		DKey:          key,
		DValue:        value.Value(),
		DGroupKeyList: value.GroupKeyList(),
		DTxnState:     value.TxnState(),
		DTValid:       value.TValid(),
		DTLease:       value.TLease(),
		DPrev:         value.Prev(),
		DLinkedLen:    value.LinkedLen(),
		DIsDeleted:    value.IsDeleted(),
		DVersion:      newVer,
	})
	if err != nil {
		return "", err
	}

	_, err = d.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName:           aws.String(d.tableName),
		Item:                av,
		ConditionExpression: aws.String("attribute_not_exists(ID)"),
	})
	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return "", errors.New(txn.VersionMismatch)
		}
		return "", err
	}

	return newVer, nil
}

func (d *DynamoDBConnection) Get(key string) (string, error) {
	if !d.hasConnected {
		return "", fmt.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	result, err := d.client.GetItem(context.Background(), &dynamodb.GetItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: key},
		},
	})
	if err != nil {
		return "", err
	}

	if result.Item == nil {
		return "", errors.New(txn.KeyNotFound)
	}

	var item KeyValueItem
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return "", err
	}

	return item.Value, nil
}

func (d *DynamoDBConnection) Put(key string, value any) error {
	if !d.hasConnected {
		return fmt.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	str := util.ToString(value)
	_, err := d.client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(d.tableName),
		Item: map[string]types.AttributeValue{
			"ID":    &types.AttributeValueMemberS{Value: key},
			"Value": &types.AttributeValueMemberS{Value: str},
		},
	})

	return err
}

func (d *DynamoDBConnection) Delete(key string) error {
	if !d.hasConnected {
		return fmt.Errorf("not connected to DynamoDB")
	}

	if oreoconfig.Debug.DebugMode {
		time.Sleep(oreoconfig.Debug.ConnAdditionalLatency)
	}

	_, err := d.client.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(d.tableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: key},
		},
	})

	return err
}

// func buildCreateTableInput(tableName string) *dynamodb.CreateTableInput {
// 	return &dynamodb.CreateTableInput{
// 		AttributeDefinitions: []types.AttributeDefinition{
// 			{
// 				AttributeName: aws.String("PK"),
// 				AttributeType: types.ScalarAttributeTypeS,
// 			},
// 			{
// 				AttributeName: aws.String("SK"),
// 				AttributeType: types.ScalarAttributeTypeS,
// 			},
// 		},
// 		KeySchema: []types.KeySchemaElement{
// 			{
// 				AttributeName: aws.String("PK"),
// 				KeyType:       types.KeyTypeHash,
// 			},
// 			{
// 				AttributeName: aws.String("SK"),
// 				KeyType:       types.KeyTypeRange,
// 			},
// 		},
// 		TableName:   aws.String(tableName),
// 		BillingMode: types.BillingModePayPerRequest,
// 	}
// }
