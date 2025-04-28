package awsmocker

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
)

// Override the default settings when using a default IMDS mock
type IMDSMockOptions struct {
	// The identity document to return
	IdentityDocument imds.InstanceIdentityDocument

	// any custom user data
	UserData string

	// if you want to override the role name that is used for EC2 creds
	RoleName string

	// Override the instance profile name
	InstanceProfileName string
}

type IMDSMockOptionFunc = func(*IMDSMockOptions)

func getDefaultImdsIdentityDocument() imds.InstanceIdentityDocument {
	return imds.InstanceIdentityDocument{
		Version:          "2017-09-30",
		InstanceID:       "i-000deadbeef",
		AccountID:        DefaultAccountId,
		Region:           DefaultRegion,
		AvailabilityZone: DefaultRegion + "a",
		InstanceType:     "t3.medium",
	}
}

// Provides an array of mocks that will provide a decent replication of the
// EC2 Instance Metadata Service
func Mock_IMDS_Common(optFns ...IMDSMockOptionFunc) []*MockedEndpoint {

	cfg := IMDSMockOptions{
		IdentityDocument:    getDefaultImdsIdentityDocument(),
		UserData:            "# awsmocker",
		RoleName:            "awsmocker_role",
		InstanceProfileName: "awsmocker-instance-profile",
	}

	for _, f := range optFns {
		f(&cfg)
	}

	mocks := make([]*MockedEndpoint, 0, 10)

	mocks = append(mocks, Mock_IMDS_IdentityDocument(func(iid *imds.InstanceIdentityDocument) {
		*iid = cfg.IdentityDocument
	}))

	for k, v := range map[string]string{
		"instance-id":         cfg.IdentityDocument.InstanceID,
		"instance-type":       cfg.IdentityDocument.InstanceType,
		"instance-life-cycle": "on-demand",
	} {
		mocks = append(mocks, Mock_IMDS_MetaData_KeyValue(k, v))
	}

	mocks = append(mocks, Mock_IMDS_UserData(cfg.UserData))
	mocks = append(mocks, Mock_IMDS_IAM_Info(cfg.InstanceProfileName))
	mocks = append(mocks, Mock_IMDS_IAM_RoleList(cfg.RoleName))
	mocks = append(mocks, Mock_IMDS_IAM_Credentials(cfg.RoleName))
	mocks = append(mocks, Mock_IMDS_API_Token())

	return mocks
}

// Provide a document to be returned, or nil to use a default one
func Mock_IMDS_IdentityDocument(optFns ...func(*imds.InstanceIdentityDocument)) *MockedEndpoint {
	doc := getDefaultImdsIdentityDocument()

	for _, f := range optFns {
		f(&doc)
	}

	return &MockedEndpoint{
		Request: &MockedRequest{
			Method:    http.MethodGet,
			Path:      "/latest/dynamic/instance-identity/document",
			IsEc2IMDS: true,
		},
		Response: &MockedResponse{
			Encoding: ResponseEncodingJSON,
			Body:     doc,
		},
	}
}

func Mock_IMDS_MetaData_KeyValue(k, v string) *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Method:    http.MethodGet,
			Path:      "/latest/meta-data/" + k,
			IsEc2IMDS: true,
		},
		Response: &MockedResponse{
			Encoding: ResponseEncodingText,
			Body:     v,
		},
	}
}

func Mock_IMDS_UserData(userData string) *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Method:    http.MethodGet,
			Path:      "/latest/user-data",
			IsEc2IMDS: true,
		},
		Response: &MockedResponse{
			Encoding: ResponseEncodingText,
			Body:     userData,
		},
	}
}

func Mock_IMDS_API_Token() *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Path:      "/latest/api/token",
			IsEc2IMDS: true,
		},
		Response: &MockedResponse{
			Encoding: ResponseEncodingText,
			Body:     "AwsMockerImdsToken",
		},
	}
}

func Mock_IMDS_IAM_Info(profileName string) *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Method:    http.MethodGet,
			Path:      "/latest/meta-data/iam/info",
			IsEc2IMDS: true,
		},
		Response: &MockedResponse{
			Encoding: ResponseEncodingJSON,
			Body: map[string]any{
				"Code":               "Success",
				"LastUpdated":        time.Now().UTC().Format(time.RFC3339),
				"InstanceProfileArn": fmt.Sprintf("arn:aws:iam::%s:instance-profile/%s", DefaultAccountId, profileName),
				"InstanceProfileId":  "AIPAABCDEFGHIJKLMN123",
			},
		},
	}
}

func Mock_IMDS_IAM_RoleList(roleName string) *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Method:    http.MethodGet,
			Path:      "/latest/meta-data/iam/security-credentials/",
			IsEc2IMDS: true,
		},
		Response: &MockedResponse{
			Encoding: ResponseEncodingText,
			Body:     roleName,
		},
	}
}

func Mock_IMDS_IAM_Credentials(roleName string) *MockedEndpoint {
	return &MockedEndpoint{
		Request: &MockedRequest{
			Method: http.MethodGet,
			Path:   "/latest/meta-data/iam/security-credentials/" + roleName,
		},
		Response: &MockedResponse{
			Encoding: ResponseEncodingJSON,
			Body: map[string]any{
				"Code":            "Success",
				"Type":            "AWS-HMAC",
				"LastUpdated":     time.Now().UTC().Format(time.RFC3339),
				"Expiration":      time.Now().UTC().Add(1 * time.Hour).Format(time.RFC3339),
				"AccessKeyID":     "FAKEKEY",
				"SecretAccessKey": "fakeSecretKEY",
				"Token":           "FAKETOKEN",
			},
		},
	}
}
