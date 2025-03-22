package api

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"
)

func checkEndpointForMeetingImportance(ctx context.Context, body ReschedulingCheckBodySchema) (bool, error) {

	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-2"))

	if err != nil {
		return false, fmt.Errorf("failed to load configuation %w", err)
	}

	// Create a new SageMaker client
	client := sagemakerruntime.NewFromConfig(cfg)

	params := sagemakerruntime.InvokeEndpointInput{
		Body:         []byte{},
		EndpointName: aws.String("sagemaker-xgboost-2025-03-20-18-18-32-151"),
		Accept:       aws.String("application/json"),
		ContentType:  aws.String("text/csv"),
	}

	res, err := client.InvokeEndpoint(ctx, &params)

	if err != nil {
		return false, fmt.Errorf("failed to invoke endpoint %w", err)
	}

	println(res)

	return false, nil
}
