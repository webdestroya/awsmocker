package awsmocker

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMockedRequest_Inspect(t *testing.T) {
	mr := &MockedRequest{
		Strict:    true,
		Hostname:  "awsmocker.local",
		Service:   "sts",
		Action:    "GetCallerIdentity",
		Body:      "somebody",
		Method:    http.MethodGet,
		Path:      "/blah/blah/some/path",
		PathRegex: regexp.MustCompile(`/version/([a-z0-9]+)/test`),
		IsEc2IMDS: true,
		Params: url.Values{
			"TestParam": []string{"thing"},
		},
	}

	result := mr.Inspect()

	require.Contains(t, result, "STRICT")
	require.Contains(t, result, "Service=sts")
	require.Contains(t, result, "Action=GetCallerIdentity")
	require.Contains(t, result, "imds=true")
	require.Contains(t, result, "Hostname=awsmocker.local")
	require.Contains(t, result, "Path=/blah/blah/some/path")
	require.Contains(t, result, "Method=GET")
	require.Contains(t, result, "Params=TestParam=thing")
	require.Contains(t, result, "Body=somebody")
}

func TestMockedRequest_matchRequest(t *testing.T) {
	type submatch struct {
		matches bool
		rr      *ReceivedRequest
	}
	tables := []struct {
		name string
		mr   *MockedRequest
		reqs []submatch
	}{
		{
			name: "SimpleServiceAction",
			mr: &MockedRequest{
				Service: "ecs",
				Action:  "CreateService",
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "UpdateService",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Service: "sts",
						Action:  "GetCallerIdentity",
					},
				},
			},
		},

		{
			name: "PathMatch",
			mr: &MockedRequest{
				Path: "/test",
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Path:     "/test",
						Hostname: "something.com",
					},
				},
				{
					matches: true,
					rr: &ReceivedRequest{
						Path:     "/test",
						Hostname: "example.com",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Path:     "/texxxst",
						Hostname: "something.com",
					},
				},
			},
		},

		{
			name: "PathRegex",
			mr: &MockedRequest{
				PathRegex: regexp.MustCompile(`/test/?.*`),
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Path: "/test",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Path: "/xtest",
					},
				},
				{
					matches: true,
					rr: &ReceivedRequest{
						Path: "/test/thing",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Path: "/yar",
					},
				},
			},
		},

		{
			name: "BaseHostnameMatch",
			mr: &MockedRequest{
				Hostname: "test.com",
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Hostname: "test.com",
						Path:     "/test",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Hostname: "www.test.com",
						Path:     "/xtest",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Path: "/xtest",
					},
				},
			},
		},

		{
			name: "BaseMethodMatch",
			mr: &MockedRequest{
				Method: http.MethodHead,
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Hostname: "test.com",
						Path:     "/test",
						HttpRequest: &http.Request{
							Method: http.MethodHead,
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Hostname: "test.com",
						Path:     "/test",
						HttpRequest: &http.Request{
							Method: http.MethodGet,
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Hostname: "test.com",
						Path:     "/test",
						HttpRequest: &http.Request{
							Method: http.MethodPost,
						},
					},
				},
			},
		},

		{
			name: "CustomMatcher",
			mr: &MockedRequest{
				Matcher: func(rr *ReceivedRequest) bool {
					return rr.Path == "/test" || rr.Hostname == "fake.com" || rr.Service == "ec2"
				},
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Path: "/test",
					},
				},
				{
					matches: true,
					rr: &ReceivedRequest{
						Path:     "/something",
						Hostname: "fake.com",
					},
				},
				{
					matches: true,
					rr: &ReceivedRequest{
						Service: "ec2",
						Action:  "TerminateInstance",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Path: "/blah",
					},
				},
			},
		},

		{
			name: "JMESpath",
			mr: &MockedRequest{
				JMESPathMatches: map[string]any{
					"thing.test[0]": "foo",
				},
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Hostname: imdsHost4,
						Path:     "/test",
						JsonPayload: map[string]any{
							"thing": map[string]any{
								"test": []string{"foo", "blah"},
							},
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Hostname: imdsHost6,
						Path:     "/test",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Hostname:    imdsHost6,
						Path:        "/test",
						JsonPayload: make(map[string]any),
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Hostname: imdsHost4,
						Path:     "/test",
						JsonPayload: map[string]any{
							"thing": map[string]any{
								"test": []string{"yar", "blah"},
							},
						},
					},
				},
			},
		},

		{
			name: "EC2Metadata",
			mr: &MockedRequest{
				IsEc2IMDS: true,
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Hostname: imdsHost4,
						Path:     "/test",
					},
				},
				{
					matches: true,
					rr: &ReceivedRequest{
						Hostname: imdsHost6,
						Path:     "/test",
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Path:     "/something",
						Hostname: "127.0.0.1",
					},
				},
			},
		},

		{
			name: "ParamsMatcher",
			mr: &MockedRequest{
				Service: "ecs",
				Action:  "CreateService",
				Params: url.Values{
					"Names.member.1": []string{"test"},
				},
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
						HttpRequest: &http.Request{
							Form: url.Values{
								"Names.member.1": []string{"test"},
							},
						},
					},
				},
				{
					matches: true,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
						HttpRequest: &http.Request{
							Form: url.Values{
								"Names.member.1":     []string{"test"},
								"Something.member.1": []string{"xxx"},
							},
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
						HttpRequest: &http.Request{
							Form: url.Values{
								"Names.member.1": []string{"test", "thing"},
							},
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
						HttpRequest: &http.Request{
							Form: url.Values{
								"Names.member.1": []string{"somethingelse"},
							},
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
						HttpRequest: &http.Request{
							Form: url.Values{
								"Something.Else.1": []string{"somethingelse"},
							},
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Service:     "ecs",
						Action:      "CreateService",
						HttpRequest: &http.Request{},
					},
				},
			},
		},

		{
			name: "StrictParamsMatcher",
			mr: &MockedRequest{
				Service: "ecs",
				Action:  "CreateService",
				Strict:  true,
				Params: url.Values{
					"Names.member.1": []string{"test"},
				},
			},
			reqs: []submatch{
				{
					matches: true,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
						HttpRequest: &http.Request{
							Form: url.Values{
								"Names.member.1": []string{"test"},
							},
						},
					},
				},
				{
					matches: false,
					rr: &ReceivedRequest{
						Service: "ecs",
						Action:  "CreateService",
						HttpRequest: &http.Request{
							Form: url.Values{
								"Names.member.1":     []string{"test"},
								"Something.member.1": []string{"xxx"},
							},
						},
					},
				},
			},
		},
	}

	for _, table := range tables {

		t.Run(table.name, func(t *testing.T) {

			for x, req := range table.reqs {
				t.Run(fmt.Sprintf("test_%d", x+1), func(t *testing.T) {
					require.Equal(t, req.matches, table.mr.matchRequest(req.rr))
				})
			}
		})
	}
}
