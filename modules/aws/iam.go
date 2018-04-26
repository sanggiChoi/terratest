package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"testing"
	"github.com/aws/aws-sdk-go/aws"
	"time"
	"github.com/gruntwork-io/terratest/modules/logger"
)

// Get the username fo the current IAM user
func GetIamCurrentUserName(t *testing.T) string {
	out, err := GetIamCurrentUserNameE(t)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Get the username fo the current IAM user
func GetIamCurrentUserNameE(t *testing.T) (string, error) {
	iamClient, err := NewIamClient(defaultRegion)
	if err != nil {
		return "", err
	}

	resp, err := iamClient.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", err
	}

	return *resp.User.UserName, nil
}

// Get the ARN for the current IAM user
func GetIamCurrentUserArn(t *testing.T) string {
	out, err := GetIamCurrentUserArnE(t)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// Get the ARN for the current IAM user
func GetIamCurrentUserArnE(t *testing.T) (string, error) {
	iamClient, err := NewIamClient(defaultRegion)
	if err != nil {
		return "", err
	}

	resp, err := iamClient.GetUser(&iam.GetUserInput{})
	if err != nil {
		return "", err
	}

	return *resp.User.Arn, nil
}

// Create an MFA device using the given IAM client
func CreateMfaDevice(t *testing.T, iamClient *iam.IAM, deviceName string) *iam.VirtualMFADevice {
	mfaDevice, err := CreateMfaDeviceE(t, iamClient, deviceName)
	if err != nil {
		t.Fatal(err)
	}
	return mfaDevice
}

// Create an MFA device using the given IAM client
func CreateMfaDeviceE(t *testing.T, iamClient *iam.IAM, deviceName string) (*iam.VirtualMFADevice, error) {
	logger.Logf(t, "Creating an MFA device called %s", deviceName)

	output, err := iamClient.CreateVirtualMFADevice(&iam.CreateVirtualMFADeviceInput{
		VirtualMFADeviceName: aws.String(deviceName),
	})
	if err != nil {
		return nil, err
	}

	if err := EnableMfaDeviceE(t, iamClient, output.VirtualMFADevice); err != nil {
		return nil, err
	}

	return output.VirtualMFADevice, nil
}

// Enable a newly created MFA Device by supplying the first two one-time passwords, so that it can be used for future
// logins by the given IAM User
func EnableMfaDevice(t *testing.T, iamClient *iam.IAM, mfaDevice *iam.VirtualMFADevice) {
	err := EnableMfaDeviceE(t, iamClient, mfaDevice)
	if err != nil {
		t.Fatal(err)
	}
}

// Enable a newly created MFA Device by supplying the first two one-time passwords, so that it can be used for future
// logins by the given IAM User
func EnableMfaDeviceE(t *testing.T, iamClient *iam.IAM, mfaDevice *iam.VirtualMFADevice) error {
	logger.Logf(t, "Enabling MFA device %s", aws.StringValue(mfaDevice.SerialNumber))

	iamUserName, err := GetIamCurrentUserArnE(t)
	if err != nil {
		return err
	}

	authCode1, err := GetTimeBasedOneTimePassword(mfaDevice)
	if err != nil {
		return err
	}

	logger.Logf(t, "Waiting 30 seconds for a new MFA Token to be generated...")
	time.Sleep(30 * time.Second)

	authCode2, err := GetTimeBasedOneTimePassword(mfaDevice)
	if err != nil {
		return err
	}

	_, err = iamClient.EnableMFADevice(&iam.EnableMFADeviceInput{
		AuthenticationCode1: aws.String(authCode1),
		AuthenticationCode2: aws.String(authCode2),
		SerialNumber:        mfaDevice.SerialNumber,
		UserName:            aws.String(iamUserName),
	})

	if err != nil {
		return err
	}

	logger.Log(t, "Waiting for MFA Device enablement to propagate.")
	time.Sleep(10 * time.Second)

	return nil
}

// Create a new IAM client
func NewIamClient(region string) (*iam.IAM, error) {
	sess, err := NewAuthenticatedSession(region)
	if err != nil {
		return nil, err
	}
	return iam.New(sess), nil
}
