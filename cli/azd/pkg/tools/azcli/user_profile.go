package azcli

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/cloud"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/azure/azure-dev/cli/azd/pkg/auth"
	"github.com/azure/azure-dev/cli/azd/pkg/graphsdk"
)

// UserProfileService allows querying for user profile information.
type UserProfileService struct {
	credentialProvider  auth.MultiTenantCredentialProvider
	graphClientFromCred func(cred azcore.TokenCredential) (*graphsdk.GraphClient, error)
}

func NewUserProfileService(
	credentialProvider auth.MultiTenantCredentialProvider,
	graphClientFromCred func(cred azcore.TokenCredential) (*graphsdk.GraphClient, error),
) *UserProfileService {
	return &UserProfileService{credentialProvider, graphClientFromCred}
}

func (user *UserProfileService) GetSignedInUserId(ctx context.Context, tenantId string) (string, error) {
	cred, err := user.credentialProvider.GetTokenCredential(ctx, tenantId)
	if err != nil {
		return "", err
	}

	graphClient, err := user.graphClientFromCred(cred)
	if err != nil {
		return "", fmt.Errorf("failed creating graph client: %w", err)
	}

	userProfile, err := graphClient.Me().Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed retrieving current user profile: %w", err)
	}

	return userProfile.Id, nil
}

func (u *UserProfileService) GetAccessToken(ctx context.Context, tenantId string) (*AzCliAccessToken, error) {
	cred, err := u.credentialProvider.GetTokenCredential(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	token, err := cred.GetToken(ctx, policy.TokenRequestOptions{
		Scopes: []string{
			fmt.Sprintf("%s/.default", cloud.AzurePublic.Services[cloud.ResourceManager].Audience),
		},
	})

	if err != nil {
		// This could happen currently if auth returned an azcli credential underneath
		if isNotLoggedInMessage(err.Error()) {
			return nil, ErrAzCliNotLoggedIn
		} else if isRefreshTokenExpiredMessage(err.Error()) {
			return nil, ErrAzCliRefreshTokenExpired
		}

		return nil, fmt.Errorf("failed retrieving access token: %w", err)
	}

	return &AzCliAccessToken{
		AccessToken: token.Token,
		ExpiresOn:   &token.ExpiresOn,
	}, nil
}
