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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID        string
	location              = "centalus"
	resourceGroupName     = "abutcher-az-vxgb6-rg"
	privateZoneName       = "sample-private-zone"
	relativeRecordSetName = "sample-relative-record-set"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	// Select user-assigned identity via its clientID.
	// Does the clientID come from a secret? OR do we create a NewManagedIdentityCredential
	// with no options to select a system-assigned identity instead?
	clientID := azidentity.ClientID("7be31448-2452-4257-a67e-24cdd7fad509")
	opts := azidentity.ManagedIdentityCredentialOptions{ID: clientID}
	cred, err := azidentity.NewManagedIdentityCredential(&opts)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	/*
		resourceGroup, err := createResourceGroup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("resources group:", *resourceGroup.ID)
	*/

	privateZone, err := createPrivateZone(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("private zone:", *privateZone.ID)

	recordSets, err := createRecordSets(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("record sets:", *recordSets.ID)

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

func createRecordSets(ctx context.Context, cred azcore.TokenCredential) (*armprivatedns.RecordSet, error) {
	recordSets, err := armprivatedns.NewRecordSetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := recordSets.CreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		armprivatedns.RecordTypeA,
		relativeRecordSetName,
		armprivatedns.RecordSet{
			Properties: &armprivatedns.RecordSetProperties{
				ARecords: []*armprivatedns.ARecord{},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.RecordSet, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
