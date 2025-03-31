// Code generated by smithy-go-codegen DO NOT EDIT.

package s3

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	s3cust "github.com/aws/aws-sdk-go-v2/service/s3/internal/customizations"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
	"time"
)

// This operation aborts a multipart upload. After a multipart upload is aborted,
// no additional parts can be uploaded using that upload ID. The storage consumed
// by any previously uploaded parts will be freed. However, if any part uploads are
// currently in progress, those part uploads might or might not succeed. As a
// result, it might be necessary to abort a given multipart upload multiple times
// in order to completely free all storage consumed by all parts.
//
// To verify that all parts have been removed and prevent getting charged for the
// part storage, you should call the [ListParts]API operation and ensure that the parts list
// is empty.
//
//   - Directory buckets - If multipart uploads in a directory bucket are in
//     progress, you can't delete the bucket until all the in-progress multipart
//     uploads are aborted or completed. To delete these in-progress multipart uploads,
//     use the ListMultipartUploads operation to list the in-progress multipart
//     uploads in the bucket and use the AbortMultipartUpload operation to abort all
//     the in-progress multipart uploads.
//
//   - Directory buckets - For directory buckets, you must make requests for this
//     API operation to the Zonal endpoint. These endpoints support
//     virtual-hosted-style requests in the format
//     https://amzn-s3-demo-bucket.s3express-zone-id.region-code.amazonaws.com/key-name
//     . Path-style requests are not supported. For more information about endpoints
//     in Availability Zones, see [Regional and Zonal endpoints for directory buckets in Availability Zones]in the Amazon S3 User Guide. For more information
//     about endpoints in Local Zones, see [Concepts for directory buckets in Local Zones]in the Amazon S3 User Guide.
//
// Permissions
//
//   - General purpose bucket permissions - For information about permissions
//     required to use the multipart upload, see [Multipart Upload and Permissions]in the Amazon S3 User Guide.
//
//   - Directory bucket permissions - To grant access to this API operation on a
//     directory bucket, we recommend that you use the [CreateSession]CreateSession API operation
//     for session-based authorization. Specifically, you grant the
//     s3express:CreateSession permission to the directory bucket in a bucket policy
//     or an IAM identity-based policy. Then, you make the CreateSession API call on
//     the bucket to obtain a session token. With the session token in your request
//     header, you can make API requests to this operation. After the session token
//     expires, you make another CreateSession API call to generate a new session
//     token for use. Amazon Web Services CLI or SDKs create session and refresh the
//     session token automatically to avoid service interruptions when a session
//     expires. For more information about authorization, see [CreateSession]CreateSession .
//
// HTTP Host header syntax  Directory buckets - The HTTP Host header syntax is
// Bucket-name.s3express-zone-id.region-code.amazonaws.com .
//
// The following operations are related to AbortMultipartUpload :
//
// [CreateMultipartUpload]
//
// [UploadPart]
//
// [CompleteMultipartUpload]
//
// [ListParts]
//
// [ListMultipartUploads]
//
// [ListParts]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListParts.html
// [Concepts for directory buckets in Local Zones]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-lzs-for-directory-buckets.html
// [UploadPart]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_UploadPart.html
// [ListMultipartUploads]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_ListMultipartUploads.html
// [CreateSession]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateSession.html
// [Multipart Upload and Permissions]: https://docs.aws.amazon.com/AmazonS3/latest/dev/mpuAndPermissions.html
// [CompleteMultipartUpload]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_CompleteMultipartUpload.html
// [Regional and Zonal endpoints for directory buckets in Availability Zones]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/endpoint-directory-buckets-AZ.html
// [CreateMultipartUpload]: https://docs.aws.amazon.com/AmazonS3/latest/API/API_CreateMultipartUpload.html
func (c *Client) AbortMultipartUpload(ctx context.Context, params *AbortMultipartUploadInput, optFns ...func(*Options)) (*AbortMultipartUploadOutput, error) {
	if params == nil {
		params = &AbortMultipartUploadInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "AbortMultipartUpload", params, optFns, c.addOperationAbortMultipartUploadMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*AbortMultipartUploadOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type AbortMultipartUploadInput struct {

	// The bucket name to which the upload was taking place.
	//
	// Directory buckets - When you use this operation with a directory bucket, you
	// must use virtual-hosted-style requests in the format
	// Bucket-name.s3express-zone-id.region-code.amazonaws.com . Path-style requests
	// are not supported. Directory bucket names must be unique in the chosen Zone
	// (Availability Zone or Local Zone). Bucket names must follow the format
	// bucket-base-name--zone-id--x-s3 (for example,
	// amzn-s3-demo-bucket--usw2-az1--x-s3 ). For information about bucket naming
	// restrictions, see [Directory bucket naming rules]in the Amazon S3 User Guide.
	//
	// Access points - When you use this action with an access point, you must provide
	// the alias of the access point in place of the bucket name or specify the access
	// point ARN. When using the access point ARN, you must direct requests to the
	// access point hostname. The access point hostname takes the form
	// AccessPointName-AccountId.s3-accesspoint.Region.amazonaws.com. When using this
	// action with an access point through the Amazon Web Services SDKs, you provide
	// the access point ARN in place of the bucket name. For more information about
	// access point ARNs, see [Using access points]in the Amazon S3 User Guide.
	//
	// Access points and Object Lambda access points are not supported by directory
	// buckets.
	//
	// S3 on Outposts - When you use this action with S3 on Outposts, you must direct
	// requests to the S3 on Outposts hostname. The S3 on Outposts hostname takes the
	// form AccessPointName-AccountId.outpostID.s3-outposts.Region.amazonaws.com . When
	// you use this action with S3 on Outposts, the destination bucket must be the
	// Outposts access point ARN or the access point alias. For more information about
	// S3 on Outposts, see [What is S3 on Outposts?]in the Amazon S3 User Guide.
	//
	// [Directory bucket naming rules]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/directory-bucket-naming-rules.html
	// [What is S3 on Outposts?]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/S3onOutposts.html
	// [Using access points]: https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-access-points.html
	//
	// This member is required.
	Bucket *string

	// Key of the object for which the multipart upload was initiated.
	//
	// This member is required.
	Key *string

	// Upload ID that identifies the multipart upload.
	//
	// This member is required.
	UploadId *string

	// The account ID of the expected bucket owner. If the account ID that you provide
	// does not match the actual owner of the bucket, the request fails with the HTTP
	// status code 403 Forbidden (access denied).
	ExpectedBucketOwner *string

	// If present, this header aborts an in progress multipart upload only if it was
	// initiated on the provided timestamp. If the initiated timestamp of the multipart
	// upload does not match the provided value, the operation returns a 412
	// Precondition Failed error. If the initiated timestamp matches or if the
	// multipart upload doesn’t exist, the operation returns a 204 Success (No Content)
	// response.
	//
	// This functionality is only supported for directory buckets.
	IfMatchInitiatedTime *time.Time

	// Confirms that the requester knows that they will be charged for the request.
	// Bucket owners need not specify this parameter in their requests. If either the
	// source or destination S3 bucket has Requester Pays enabled, the requester will
	// pay for corresponding charges to copy the object. For information about
	// downloading objects from Requester Pays buckets, see [Downloading Objects in Requester Pays Buckets]in the Amazon S3 User
	// Guide.
	//
	// This functionality is not supported for directory buckets.
	//
	// [Downloading Objects in Requester Pays Buckets]: https://docs.aws.amazon.com/AmazonS3/latest/dev/ObjectsinRequesterPaysBuckets.html
	RequestPayer types.RequestPayer

	noSmithyDocumentSerde
}

func (in *AbortMultipartUploadInput) bindEndpointParams(p *EndpointParameters) {

	p.Bucket = in.Bucket
	p.Key = in.Key

}

type AbortMultipartUploadOutput struct {

	// If present, indicates that the requester was successfully charged for the
	// request.
	//
	// This functionality is not supported for directory buckets.
	RequestCharged types.RequestCharged

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationAbortMultipartUploadMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsRestxml_serializeOpAbortMultipartUpload{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestxml_deserializeOpAbortMultipartUpload{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "AbortMultipartUpload"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addClientRequestID(stack); err != nil {
		return err
	}
	if err = addComputeContentLength(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addComputePayloadSHA256(stack); err != nil {
		return err
	}
	if err = addRetry(stack, options); err != nil {
		return err
	}
	if err = addRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = addRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addSpanRetryLoop(stack, options); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addPutBucketContextMiddleware(stack); err != nil {
		return err
	}
	if err = addTimeOffsetBuild(stack, c); err != nil {
		return err
	}
	if err = addUserAgentRetryMode(stack, options); err != nil {
		return err
	}
	if err = addIsExpressUserAgent(stack); err != nil {
		return err
	}
	if err = addCredentialSource(stack, options); err != nil {
		return err
	}
	if err = addOpAbortMultipartUploadValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opAbortMultipartUpload(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addMetadataRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addRecursionDetection(stack); err != nil {
		return err
	}
	if err = addAbortMultipartUploadUpdateEndpoint(stack, options); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = v4.AddContentSHA256HeaderMiddleware(stack); err != nil {
		return err
	}
	if err = disableAcceptEncodingGzip(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	if err = addSerializeImmutableHostnameBucketMiddleware(stack, options); err != nil {
		return err
	}
	if err = addSpanInitializeStart(stack); err != nil {
		return err
	}
	if err = addSpanInitializeEnd(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestStart(stack); err != nil {
		return err
	}
	if err = addSpanBuildRequestEnd(stack); err != nil {
		return err
	}
	return nil
}

func (v *AbortMultipartUploadInput) bucket() (string, bool) {
	if v.Bucket == nil {
		return "", false
	}
	return *v.Bucket, true
}

func newServiceMetadataMiddleware_opAbortMultipartUpload(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "AbortMultipartUpload",
	}
}

// getAbortMultipartUploadBucketMember returns a pointer to string denoting a
// provided bucket member valueand a boolean indicating if the input has a modeled
// bucket name,
func getAbortMultipartUploadBucketMember(input interface{}) (*string, bool) {
	in := input.(*AbortMultipartUploadInput)
	if in.Bucket == nil {
		return nil, false
	}
	return in.Bucket, true
}
func addAbortMultipartUploadUpdateEndpoint(stack *middleware.Stack, options Options) error {
	return s3cust.UpdateEndpoint(stack, s3cust.UpdateEndpointOptions{
		Accessor: s3cust.UpdateEndpointParameterAccessor{
			GetBucketFromInput: getAbortMultipartUploadBucketMember,
		},
		UsePathStyle:                   options.UsePathStyle,
		UseAccelerate:                  options.UseAccelerate,
		SupportsAccelerate:             true,
		TargetS3ObjectLambda:           false,
		EndpointResolver:               options.EndpointResolver,
		EndpointResolverOptions:        options.EndpointOptions,
		UseARNRegion:                   options.UseARNRegion,
		DisableMultiRegionAccessPoints: options.DisableMultiRegionAccessPoints,
	})
}
