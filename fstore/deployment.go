package fstore

import (
	"fmt"
	"net/url"

	"github.com/blessnetwork/b7s/models/bls"
)

// updateDeploymentInfo will add some missing information to the deployment info,
// based on the info found in the manifest URL.
func updateDeploymentInfo(manifest *bls.FunctionManifest, manifestAddress string) error {

	// Parse the deployment address.
	deploymentURL, err := url.Parse(manifest.Runtime.URL)
	if err != nil {
		return fmt.Errorf("could not parse manifest runtime URL: %w", err)
	}

	// Parse the provided manifest address.
	manifestURL, err := url.Parse(manifestAddress)
	if err != nil {
		return fmt.Errorf("could not parse manifest URL: %w", err)
	}

	// Fill in missing address data using the manifest address info.
	if deploymentURL.Host == "" {
		deploymentURL.Host = manifestURL.Host
	}
	if deploymentURL.Scheme == "" {
		deploymentURL.Scheme = manifestURL.Scheme
	}

	manifest.Deployment = bls.Deployment{
		URI:      deploymentURL.String(),
		Checksum: manifest.Runtime.Checksum,
	}

	return nil
}
