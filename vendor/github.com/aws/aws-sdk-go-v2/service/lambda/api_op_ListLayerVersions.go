// Code generated by smithy-go-codegen DO NOT EDIT.

package lambda

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Lists the versions of an Lambda layer
// (https://docs.aws.amazon.com/lambda/latest/dg/configuration-layers.html).
// Versions that have been deleted aren't listed. Specify a runtime identifier
// (https://docs.aws.amazon.com/lambda/latest/dg/lambda-runtimes.html) to list only
// versions that indicate that they're compatible with that runtime. Specify a
// compatible architecture to include only layer versions that are compatible with
// that architecture.
func (c *Client) ListLayerVersions(ctx context.Context, params *ListLayerVersionsInput, optFns ...func(*Options)) (*ListLayerVersionsOutput, error) {
	if params == nil {
		params = &ListLayerVersionsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListLayerVersions", params, optFns, c.addOperationListLayerVersionsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListLayerVersionsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListLayerVersionsInput struct {

	// The name or Amazon Resource Name (ARN) of the layer.
	//
	// This member is required.
	LayerName *string

	// The compatible instruction set architecture
	// (https://docs.aws.amazon.com/lambda/latest/dg/foundation-arch.html).
	CompatibleArchitecture types.Architecture

	// A runtime identifier. For example, go1.x.
	CompatibleRuntime types.Runtime

	// A pagination token returned by a previous call.
	Marker *string

	// The maximum number of versions to return.
	MaxItems *int32

	noSmithyDocumentSerde
}

type ListLayerVersionsOutput struct {

	// A list of versions.
	LayerVersions []types.LayerVersionsListItem

	// A pagination token returned when the response doesn't contain all versions.
	NextMarker *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationListLayerVersionsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestjson1_serializeOpListLayerVersions{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestjson1_deserializeOpListLayerVersions{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpListLayerVersionsValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListLayerVersions(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

// ListLayerVersionsAPIClient is a client that implements the ListLayerVersions
// operation.
type ListLayerVersionsAPIClient interface {
	ListLayerVersions(context.Context, *ListLayerVersionsInput, ...func(*Options)) (*ListLayerVersionsOutput, error)
}

var _ ListLayerVersionsAPIClient = (*Client)(nil)

// ListLayerVersionsPaginatorOptions is the paginator options for ListLayerVersions
type ListLayerVersionsPaginatorOptions struct {
	// The maximum number of versions to return.
	Limit int32

	// Set to true if pagination should stop if the service returns a pagination token
	// that matches the most recent token provided to the service.
	StopOnDuplicateToken bool
}

// ListLayerVersionsPaginator is a paginator for ListLayerVersions
type ListLayerVersionsPaginator struct {
	options   ListLayerVersionsPaginatorOptions
	client    ListLayerVersionsAPIClient
	params    *ListLayerVersionsInput
	nextToken *string
	firstPage bool
}

// NewListLayerVersionsPaginator returns a new ListLayerVersionsPaginator
func NewListLayerVersionsPaginator(client ListLayerVersionsAPIClient, params *ListLayerVersionsInput, optFns ...func(*ListLayerVersionsPaginatorOptions)) *ListLayerVersionsPaginator {
	if params == nil {
		params = &ListLayerVersionsInput{}
	}

	options := ListLayerVersionsPaginatorOptions{}
	if params.MaxItems != nil {
		options.Limit = *params.MaxItems
	}

	for _, fn := range optFns {
		fn(&options)
	}

	return &ListLayerVersionsPaginator{
		options:   options,
		client:    client,
		params:    params,
		firstPage: true,
	}
}

// HasMorePages returns a boolean indicating whether more pages are available
func (p *ListLayerVersionsPaginator) HasMorePages() bool {
	return p.firstPage || p.nextToken != nil
}

// NextPage retrieves the next ListLayerVersions page.
func (p *ListLayerVersionsPaginator) NextPage(ctx context.Context, optFns ...func(*Options)) (*ListLayerVersionsOutput, error) {
	if !p.HasMorePages() {
		return nil, fmt.Errorf("no more pages available")
	}

	params := *p.params
	params.Marker = p.nextToken

	var limit *int32
	if p.options.Limit > 0 {
		limit = &p.options.Limit
	}
	params.MaxItems = limit

	result, err := p.client.ListLayerVersions(ctx, &params, optFns...)
	if err != nil {
		return nil, err
	}
	p.firstPage = false

	prevToken := p.nextToken
	p.nextToken = result.NextMarker

	if p.options.StopOnDuplicateToken && prevToken != nil && p.nextToken != nil && *prevToken == *p.nextToken {
		p.nextToken = nil
	}

	return result, nil
}

func newServiceMetadataMiddleware_opListLayerVersions(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "lambda",
		OperationName: "ListLayerVersions",
	}
}
