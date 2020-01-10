package auth

import (
	"fmt"
	"strings"

	"github.com/jenkins-x/jx/pkg/log"
	"github.com/jenkins-x/jx/pkg/secreturl"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/yaml"
)

const (
	labelAuthConfig      = "jenkins.io/config-type"
	labelAuthConfigValue = "auth"
)

// LoadConfig loads the auth config from a ConfigMap which stores in its
// data with a key equal with the secretName, also it resolves any secrets
// URIs by fetching their secret data from vault.
func (c *ConfigMapVaultConfigHandler) LoadConfig() (*AuthConfig, error) {
	selector := fmt.Sprintf("%s=%s", labelAuthConfig, labelAuthConfigValue)
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	log.Logger().Warnf("LOOKING FOR CONFIG MAP WITH SELECTOR %s", selector)
	configList, err := c.configMapClient.List(listOptions)
	if err != nil || configList == nil {
		return nil, errors.Wrapf(err, "looking for configmaps with label %s", selector)
	}
	if len(configList.Items) == 0 {
		return nil, fmt.Errorf("no configmap with label %s found", selector)
	}
	log.Logger().Warnf("FOUND CONFIG MAPS, LOOKING FOR ONE WITH %s", c.secretName)
	var data string
	for _, config := range configList.Items {
		if d, ok := config.Data[c.secretName]; ok {
			data = d
			break
		}
	}
	if data == "" {
		return nil, fmt.Errorf("no configmap with label %s found with data key %s",
			selector, c.secretName)
	}
	log.Logger().Warnf("FOUND ONE, NOT LOGGING ITS CONTENTS")
	if data, err = c.secretURLClient.ReplaceURIs(data); err != nil {
		return nil, errors.Wrapf(err, "replacing the secrets in auth config %q from vault", c.secretName)
	}

	var config AuthConfig
	if err := yaml.Unmarshal([]byte(data), &config); err != nil {
		return nil, errors.Wrapf(err, "unmarshaling auth config %q", c.secretName)
	}
	var serverNames []string
	for _, s := range config.Servers {
		serverNames = append(serverNames, s.Name)
	}
	log.Logger().Warnf("GOT IT UNMARSHALED HERE ARE SOME SERVERS %s", strings.Join(serverNames, ", "))
	return &config, nil
}

// SaveConfig should save config but we keep this read-only to avoid
// overwriting the existing values configure during installation.
func (c *ConfigMapVaultConfigHandler) SaveConfig(config *AuthConfig) error {
	return nil
}

// NewConfigMapVaultConfigHandler creates a new configmap/vault config handler
func NewConfigMapVaultConfigHandler(secretName string, configMapClient v1.ConfigMapInterface,
	vaultClient secreturl.Client) ConfigMapVaultConfigHandler {
	return ConfigMapVaultConfigHandler{
		secretName:      secretName,
		configMapClient: configMapClient,
		secretURLClient: vaultClient,
	}
}

// NewConfigmapVaultAuthConfigService creates a new config service that load the config from
// a configmap and resolve the secrets URIs from vault
func NewConfigmapVaultAuthConfigService(secretName string, configMapClient v1.ConfigMapInterface,
	secretURLClient secreturl.Client) ConfigService {
	handler := NewConfigMapVaultConfigHandler(secretName, configMapClient, secretURLClient)
	return NewAuthConfigService(&handler)
}

// IsConfigMapVaultAuth checks if is able to find any auth config in a config map
func IsConfigMapVaultAuth(configMapClient v1.ConfigMapInterface) bool {
	selector := fmt.Sprintf("%s=%s", labelAuthConfig, labelAuthConfigValue)
	listOptions := metav1.ListOptions{
		LabelSelector: selector,
	}
	configList, err := configMapClient.List(listOptions)
	if err != nil || configList == nil {
		return false
	}
	if len(configList.Items) == 0 {
		return false
	}
	return true
}
