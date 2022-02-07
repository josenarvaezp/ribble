// Code generated by smithy-go-codegen DO NOT EDIT.

package types

import (
	"fmt"
	smithy "github.com/aws/smithy-go"
)

// The specified code signing configuration does not exist.
type CodeSigningConfigNotFoundException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *CodeSigningConfigNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *CodeSigningConfigNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *CodeSigningConfigNotFoundException) ErrorCode() string {
	return "CodeSigningConfigNotFoundException"
}
func (e *CodeSigningConfigNotFoundException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// You have exceeded your maximum total code size per account. Learn more
// (https://docs.aws.amazon.com/lambda/latest/dg/limits.html)
type CodeStorageExceededException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *CodeStorageExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *CodeStorageExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *CodeStorageExceededException) ErrorCode() string             { return "CodeStorageExceededException" }
func (e *CodeStorageExceededException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The code signature failed one or more of the validation checks for signature
// mismatch or expiry, and the code signing policy is set to ENFORCE. Lambda blocks
// the deployment.
type CodeVerificationFailedException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *CodeVerificationFailedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *CodeVerificationFailedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *CodeVerificationFailedException) ErrorCode() string {
	return "CodeVerificationFailedException"
}
func (e *CodeVerificationFailedException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Need additional permissions to configure VPC settings.
type EC2AccessDeniedException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *EC2AccessDeniedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *EC2AccessDeniedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *EC2AccessDeniedException) ErrorCode() string             { return "EC2AccessDeniedException" }
func (e *EC2AccessDeniedException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda was throttled by Amazon EC2 during Lambda function initialization using
// the execution role provided for the Lambda function.
type EC2ThrottledException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *EC2ThrottledException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *EC2ThrottledException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *EC2ThrottledException) ErrorCode() string             { return "EC2ThrottledException" }
func (e *EC2ThrottledException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda received an unexpected EC2 client exception while setting up for the
// Lambda function.
type EC2UnexpectedException struct {
	Message *string

	Type         *string
	EC2ErrorCode *string

	noSmithyDocumentSerde
}

func (e *EC2UnexpectedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *EC2UnexpectedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *EC2UnexpectedException) ErrorCode() string             { return "EC2UnexpectedException" }
func (e *EC2UnexpectedException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// An error occurred when reading from or writing to a connected file system.
type EFSIOException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *EFSIOException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *EFSIOException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *EFSIOException) ErrorCode() string             { return "EFSIOException" }
func (e *EFSIOException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The function couldn't make a network connection to the configured file system.
type EFSMountConnectivityException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *EFSMountConnectivityException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *EFSMountConnectivityException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *EFSMountConnectivityException) ErrorCode() string             { return "EFSMountConnectivityException" }
func (e *EFSMountConnectivityException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The function couldn't mount the configured file system due to a permission or
// configuration issue.
type EFSMountFailureException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *EFSMountFailureException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *EFSMountFailureException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *EFSMountFailureException) ErrorCode() string             { return "EFSMountFailureException" }
func (e *EFSMountFailureException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The function was able to make a network connection to the configured file
// system, but the mount operation timed out.
type EFSMountTimeoutException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *EFSMountTimeoutException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *EFSMountTimeoutException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *EFSMountTimeoutException) ErrorCode() string             { return "EFSMountTimeoutException" }
func (e *EFSMountTimeoutException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// Lambda was not able to create an elastic network interface in the VPC, specified
// as part of Lambda function configuration, because the limit for network
// interfaces has been reached.
type ENILimitReachedException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *ENILimitReachedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ENILimitReachedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ENILimitReachedException) ErrorCode() string             { return "ENILimitReachedException" }
func (e *ENILimitReachedException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// The code signature failed the integrity check. Lambda always blocks deployment
// if the integrity check fails, even if code signing policy is set to WARN.
type InvalidCodeSignatureException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *InvalidCodeSignatureException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidCodeSignatureException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidCodeSignatureException) ErrorCode() string             { return "InvalidCodeSignatureException" }
func (e *InvalidCodeSignatureException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// One of the parameters in the request is invalid.
type InvalidParameterValueException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *InvalidParameterValueException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidParameterValueException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidParameterValueException) ErrorCode() string             { return "InvalidParameterValueException" }
func (e *InvalidParameterValueException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The request body could not be parsed as JSON.
type InvalidRequestContentException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *InvalidRequestContentException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidRequestContentException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidRequestContentException) ErrorCode() string             { return "InvalidRequestContentException" }
func (e *InvalidRequestContentException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The runtime or runtime version specified is not supported.
type InvalidRuntimeException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *InvalidRuntimeException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidRuntimeException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidRuntimeException) ErrorCode() string             { return "InvalidRuntimeException" }
func (e *InvalidRuntimeException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// The Security Group ID provided in the Lambda function VPC configuration is
// invalid.
type InvalidSecurityGroupIDException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *InvalidSecurityGroupIDException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidSecurityGroupIDException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidSecurityGroupIDException) ErrorCode() string {
	return "InvalidSecurityGroupIDException"
}
func (e *InvalidSecurityGroupIDException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// The Subnet ID provided in the Lambda function VPC configuration is invalid.
type InvalidSubnetIDException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *InvalidSubnetIDException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidSubnetIDException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidSubnetIDException) ErrorCode() string             { return "InvalidSubnetIDException" }
func (e *InvalidSubnetIDException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda could not unzip the deployment package.
type InvalidZipFileException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *InvalidZipFileException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *InvalidZipFileException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *InvalidZipFileException) ErrorCode() string             { return "InvalidZipFileException" }
func (e *InvalidZipFileException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda was unable to decrypt the environment variables because KMS access was
// denied. Check the Lambda function's KMS permissions.
type KMSAccessDeniedException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *KMSAccessDeniedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *KMSAccessDeniedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *KMSAccessDeniedException) ErrorCode() string             { return "KMSAccessDeniedException" }
func (e *KMSAccessDeniedException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda was unable to decrypt the environment variables because the KMS key used
// is disabled. Check the Lambda function's KMS key settings.
type KMSDisabledException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *KMSDisabledException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *KMSDisabledException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *KMSDisabledException) ErrorCode() string             { return "KMSDisabledException" }
func (e *KMSDisabledException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda was unable to decrypt the environment variables because the KMS key used
// is in an invalid state for Decrypt. Check the function's KMS key settings.
type KMSInvalidStateException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *KMSInvalidStateException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *KMSInvalidStateException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *KMSInvalidStateException) ErrorCode() string             { return "KMSInvalidStateException" }
func (e *KMSInvalidStateException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda was unable to decrypt the environment variables because the KMS key was
// not found. Check the function's KMS key settings.
type KMSNotFoundException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *KMSNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *KMSNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *KMSNotFoundException) ErrorCode() string             { return "KMSNotFoundException" }
func (e *KMSNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// The permissions policy for the resource is too large. Learn more
// (https://docs.aws.amazon.com/lambda/latest/dg/limits.html)
type PolicyLengthExceededException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *PolicyLengthExceededException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *PolicyLengthExceededException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *PolicyLengthExceededException) ErrorCode() string             { return "PolicyLengthExceededException" }
func (e *PolicyLengthExceededException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The RevisionId provided does not match the latest RevisionId for the Lambda
// function or alias. Call the GetFunction or the GetAlias API to retrieve the
// latest RevisionId for your resource.
type PreconditionFailedException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *PreconditionFailedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *PreconditionFailedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *PreconditionFailedException) ErrorCode() string             { return "PreconditionFailedException" }
func (e *PreconditionFailedException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The specified configuration does not exist.
type ProvisionedConcurrencyConfigNotFoundException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *ProvisionedConcurrencyConfigNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ProvisionedConcurrencyConfigNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ProvisionedConcurrencyConfigNotFoundException) ErrorCode() string {
	return "ProvisionedConcurrencyConfigNotFoundException"
}
func (e *ProvisionedConcurrencyConfigNotFoundException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultClient
}

// The request payload exceeded the Invoke request body JSON input limit. For more
// information, see Limits
// (https://docs.aws.amazon.com/lambda/latest/dg/limits.html).
type RequestTooLargeException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *RequestTooLargeException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *RequestTooLargeException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *RequestTooLargeException) ErrorCode() string             { return "RequestTooLargeException" }
func (e *RequestTooLargeException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The resource already exists, or another operation is in progress.
type ResourceConflictException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *ResourceConflictException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceConflictException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceConflictException) ErrorCode() string             { return "ResourceConflictException" }
func (e *ResourceConflictException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The operation conflicts with the resource's availability. For example, you
// attempted to update an EventSource Mapping in CREATING, or tried to delete a
// EventSource mapping currently in the UPDATING state.
type ResourceInUseException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *ResourceInUseException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceInUseException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceInUseException) ErrorCode() string             { return "ResourceInUseException" }
func (e *ResourceInUseException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The resource specified in the request does not exist.
type ResourceNotFoundException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *ResourceNotFoundException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceNotFoundException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceNotFoundException) ErrorCode() string             { return "ResourceNotFoundException" }
func (e *ResourceNotFoundException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The function is inactive and its VPC connection is no longer available. Wait for
// the VPC connection to reestablish and try again.
type ResourceNotReadyException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *ResourceNotReadyException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ResourceNotReadyException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ResourceNotReadyException) ErrorCode() string             { return "ResourceNotReadyException" }
func (e *ResourceNotReadyException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// The Lambda service encountered an internal error.
type ServiceException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *ServiceException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *ServiceException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *ServiceException) ErrorCode() string             { return "ServiceException" }
func (e *ServiceException) ErrorFault() smithy.ErrorFault { return smithy.FaultServer }

// Lambda was not able to set up VPC access for the Lambda function because one or
// more configured subnets has no available IP addresses.
type SubnetIPAddressLimitReachedException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *SubnetIPAddressLimitReachedException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *SubnetIPAddressLimitReachedException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *SubnetIPAddressLimitReachedException) ErrorCode() string {
	return "SubnetIPAddressLimitReachedException"
}
func (e *SubnetIPAddressLimitReachedException) ErrorFault() smithy.ErrorFault {
	return smithy.FaultServer
}

// The request throughput limit was exceeded.
type TooManyRequestsException struct {
	Message *string

	RetryAfterSeconds *string
	Type              *string
	Reason            ThrottleReason

	noSmithyDocumentSerde
}

func (e *TooManyRequestsException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *TooManyRequestsException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *TooManyRequestsException) ErrorCode() string             { return "TooManyRequestsException" }
func (e *TooManyRequestsException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }

// The content type of the Invoke request body is not JSON.
type UnsupportedMediaTypeException struct {
	Message *string

	Type *string

	noSmithyDocumentSerde
}

func (e *UnsupportedMediaTypeException) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorCode(), e.ErrorMessage())
}
func (e *UnsupportedMediaTypeException) ErrorMessage() string {
	if e.Message == nil {
		return ""
	}
	return *e.Message
}
func (e *UnsupportedMediaTypeException) ErrorCode() string             { return "UnsupportedMediaTypeException" }
func (e *UnsupportedMediaTypeException) ErrorFault() smithy.ErrorFault { return smithy.FaultClient }
