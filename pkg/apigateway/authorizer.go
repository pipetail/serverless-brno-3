package apigw

type APIGatewayV2CustomAuthorizerRequest struct {
	Version               string            `json:"version"`
	Type                  string            `json:"type"`
	RouteARN              string            `json:"routeArn"`
	IdentitySource        []string          `json:"identitySource"`
	RouteKey              string            `json:"routeKey"`
	RawPath               string            `json:"rawPath"`
	RawQueryString        string            `json:"rawQueryString"`
	Cookies               []string          `json:"cookies"`
	Headers               map[string]string `json:"headers"`
	QueryStringParameters map[string]string `json:"queryStringParameters"`
	PathParameters        map[string]string `json:"pathParameters"`
	StageVariables        map[string]string `json:"stageVariables"`
}

type APIGatewayV2CustomAuthorizerResponse struct {
	PrincipalId    string                                             `json:"principalId"`
	Context        map[string]string                                  `json:"context"`
	PolicyDocument APIGatewayV2CustomAuthorizerResponsePolicyDocument `json:"policyDocument"`
}

type APIGatewayV2CustomAuthorizerResponsePolicyDocument struct {
	Version   string                                                `json:"Version"`
	Statement []APIGatewayV2CustomAuthorizerResponsePolicyStatement `json:"Statement"`
}

type APIGatewayV2CustomAuthorizerResponsePolicyStatement struct {
	Action   string `json:"Action"`
	Effect   string `json:"Effect"`
	Resource string `json:"Resource"`
}

func AuthorizerDeny(arn string) APIGatewayV2CustomAuthorizerResponse {
	return APIGatewayV2CustomAuthorizerResponse{
		PrincipalId: "none",
		PolicyDocument: APIGatewayV2CustomAuthorizerResponsePolicyDocument{
			Version: "2012-10-17",
			Statement: []APIGatewayV2CustomAuthorizerResponsePolicyStatement{
				{
					Action:   "execute-api:Invoke",
					Effect:   "Deny",
					Resource: arn,
				},
			},
		},
	}
}

func AuthorizerAllow(arn string, user string) APIGatewayV2CustomAuthorizerResponse {
	return APIGatewayV2CustomAuthorizerResponse{
		PrincipalId: user,
		PolicyDocument: APIGatewayV2CustomAuthorizerResponsePolicyDocument{
			Version: "2012-10-17",
			Statement: []APIGatewayV2CustomAuthorizerResponsePolicyStatement{
				{
					Action:   "execute-api:Invoke",
					Effect:   "Allow",
					Resource: arn,
				},
			},
		},
		Context: map[string]string{},
	}
}
