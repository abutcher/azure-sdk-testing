// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
)

var (
	subscriptionID        string
	clientID              string
	location              string
	resourceGroupName     string
	privateZoneName       string
	relativeRecordSetName = ""
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	clientID = os.Getenv("AZURE_CLIENT_ID")
	if len(clientID) == 0 {
		log.Fatal("AZURE_CLIENT_ID is not set.")
	}

	location = os.Getenv("AZURE_LOCATION")
	if len(location) == 0 {
		log.Fatal("AZURE_LOCATION is not set.")
	}

	resourceGroupName = os.Getenv("AZURE_RESOURCEGROUP_NAME")
	if len(resourceGroupName) == 0 {
		log.Fatal("AZURE_RESOURCEGROUP_NAME is not set.")
	}

	privateZoneName = os.Getenv("AZURE_PRIVATE_DNSZONE")
	if len(privateZoneName) == 0 {
		log.Fatal("AZURE_PRIVATE_DNSZONE is not set.")
	}

	// Select user-assigned identity via its clientID.
	// Does the clientID come from a secret?
	clientID := azidentity.ClientID(clientID)
	opts := azidentity.ManagedIdentityCredentialOptions{ID: clientID}
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	privateZone, err := createPrivateZone(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("private zone:", *privateZone.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createPrivateZone(ctx context.Context, cred azcore.TokenCredential) (*armprivatedns.PrivateZone, error) {
	privateZonesClient, err := armprivatedns.NewPrivateZonesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollersResp, err := privateZonesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		armprivatedns.PrivateZone{
			Location: to.Ptr(location),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollersResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.PrivateZone, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	privateZonesClient, err := armprivatedns.NewPrivateZonesClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollersResp, err := privateZonesClient.BeginDelete(
		ctx,
		resourceGroupName,
		privateZoneName,
		nil,
	)

	_, err = pollersResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
