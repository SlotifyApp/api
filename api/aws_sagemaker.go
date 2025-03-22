package api

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"
)

type EndpointRequestParams struct {
	newMeetingTitle     string
	newMeetingAttendees int
	newMeetingDuration  string
	oldMeetingTitle     string
	oldMeetingAttendees int
	oldMeetingDuration  string
}

func checkEndpointForMeetingImportance(ctx context.Context, body EndpointRequestParams) (bool, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion("eu-west-2"))
	if err != nil {
		return false, fmt.Errorf("failed to load configuation %w", err)
	}
	// "[1,('Weekly Meeting', 2,'90 minutes'),('Daily', 2,'90 minutes')]"
	modelInput := fmt.Sprintf("[1,('%s',%d,'%s'),('%s',%d,'%s')]", body.newMeetingTitle,
		body.newMeetingAttendees, body.newMeetingDuration, body.oldMeetingTitle,
		body.oldMeetingAttendees, body.oldMeetingDuration)

	// Create a new SageMaker client
	client := sagemakerruntime.NewFromConfig(cfg)

	params := sagemakerruntime.InvokeEndpointInput{
		Body:         []byte(modelInput),
		EndpointName: aws.String("sagemaker-xgboost-2025-03-20-18-18-32-151"),
		Accept:       aws.String("application/json"),
		ContentType:  aws.String("text/csv"),
	}

	res, err := client.InvokeEndpoint(ctx, &params)
	if err != nil {
		return false, fmt.Errorf("failed to invoke endpoint %w", err)
	}

	//nolint: forbidigo // need it for debugging
	println(res.Body)

	if res.Body != nil {
		return true, nil
	}

	return false, nil
}
