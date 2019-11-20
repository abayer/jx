package testhelpers

import (
	"testing"

	"github.com/jenkins-x/jx/pkg/cloud"
	"github.com/jenkins-x/jx/pkg/config"
	"github.com/jenkins-x/jx/pkg/gits"
	"github.com/stretchr/testify/assert"
)

// Helpers for tests around requirements

// MergeTestSpec defines a test case for use in testing of merging requirements
type MergeTestSpec struct {
	Name           string
	Original       *config.RequirementsConfig
	Changed        *config.RequirementsConfig
	ValidationFunc func(orig *config.RequirementsConfig, ch *config.RequirementsConfig)
}

// GetMergeTestCases returns a standard set of test cases for use in testing of merging requirements
func GetMergeTestCases(t *testing.T) []MergeTestSpec {
	return []MergeTestSpec{
		{
			Name: "Merge Cluster Config Test",
			Original: &config.RequirementsConfig{
				Cluster: config.ClusterConfig{
					EnvironmentGitOwner:  "owner",
					EnvironmentGitPublic: false,
					GitPublic:            false,
					Provider:             cloud.GKE,
					Namespace:            "jx",
					ProjectID:            "project-id",
					ClusterName:          "cluster-name",
					Region:               "region",
					GitKind:              gits.KindGitHub,
					GitServer:            gits.KindGitHub,
				},
			},
			Changed: &config.RequirementsConfig{
				Cluster: config.ClusterConfig{
					EnvironmentGitPublic: true,
					GitPublic:            true,
					Provider:             cloud.GKE,
				},
			},
			ValidationFunc: func(orig *config.RequirementsConfig, ch *config.RequirementsConfig) {
				assert.True(t, orig.Cluster.EnvironmentGitPublic == ch.Cluster.EnvironmentGitPublic &&
					orig.Cluster.GitPublic == ch.Cluster.GitPublic &&
					orig.Cluster.ProjectID != ch.Cluster.ProjectID, "ClusterConfig validation failed")
			},
		},
		{
			Name: "Merge EnvironmentConfig slices Test",
			Original: &config.RequirementsConfig{
				Environments: []config.EnvironmentConfig{
					{
						Key:        "dev",
						Repository: "repo",
					},
					{
						Key: "production",
						Ingress: config.IngressConfig{
							Domain: "domain",
						},
					},
					{
						Key: "staging",
						Ingress: config.IngressConfig{
							Domain: "domainStaging",
							TLS: config.TLSConfig{
								Email: "email",
							},
						},
					},
				},
			},
			Changed: &config.RequirementsConfig{
				Environments: []config.EnvironmentConfig{
					{
						Key:   "dev",
						Owner: "owner",
					},
					{
						Key: "production",
						Ingress: config.IngressConfig{
							CloudDNSSecretName: "secret",
						},
					},
					{
						Key: "staging",
						Ingress: config.IngressConfig{
							Domain:          "newDomain",
							DomainIssuerURL: "issuer",
							TLS: config.TLSConfig{
								Enabled: true,
							},
						},
					},
					{
						Key: "ns2",
					},
				},
			},
			ValidationFunc: func(orig *config.RequirementsConfig, ch *config.RequirementsConfig) {
				assert.True(t, len(orig.Environments) == len(ch.Environments), "the environment slices should be of the same len")
				// -- Dev Environment's asserts
				devOrig, devCh := orig.Environments[0], ch.Environments[0]
				assert.True(t, devOrig.Owner == devCh.Owner && devOrig.Repository != devCh.Repository,
					"the dev environment should've been merged correctly")
				// -- Production Environment's asserts
				prOrig, prCh := orig.Environments[1], ch.Environments[1]
				assert.True(t, prOrig.Ingress.Domain == "domain" &&
					prOrig.Ingress.CloudDNSSecretName == prCh.Ingress.CloudDNSSecretName,
					"the production environment should've been merged correctly")
				// -- Staging Environmnet's asserts
				stgOrig, stgCh := orig.Environments[2], ch.Environments[2]
				assert.True(t, stgOrig.Ingress.Domain == stgCh.Ingress.Domain &&
					stgOrig.Ingress.TLS.Email == "email" && stgOrig.Ingress.TLS.Enabled == stgCh.Ingress.TLS.Enabled,
					"the staging environment should've been merged correctly")
			},
		},
		{
			Name: "Merge StorageConfig test",
			Original: &config.RequirementsConfig{
				Storage: config.StorageConfig{
					Logs: config.StorageEntryConfig{
						Enabled: true,
						URL:     "value1",
					},
					Reports: config.StorageEntryConfig{},
					Repository: config.StorageEntryConfig{
						Enabled: true,
						URL:     "value3",
					},
				},
			},
			Changed: &config.RequirementsConfig{
				Storage: config.StorageConfig{
					Reports: config.StorageEntryConfig{
						Enabled: true,
						URL:     "",
					},
				},
			},
			ValidationFunc: func(orig *config.RequirementsConfig, ch *config.RequirementsConfig) {
				assert.True(t, orig.Storage.Logs.Enabled && orig.Storage.Logs.URL == "value1" &&
					orig.Storage.Repository.Enabled && orig.Storage.Repository.URL == "value3" &&
					orig.Storage.Reports.Enabled == ch.Storage.Reports.Enabled,
					"The storage configuration should've been merged correctly")
			},
		},
	}
}
