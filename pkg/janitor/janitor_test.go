package janitor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	apiv1 "k8s.io/client-go/pkg/api/v1"
)

func TestConfigMapToClusterSettings(t *testing.T) {
	cnf1 := apiv1.Secret{
		Data: map[string][]byte{
			"log-index-prefix":            []byte("test-prefix"),
			"log-storage-lifetime":        []byte("3333"),
			"monitoring-storage-lifetime": []byte("2222"),
		},
	}
	expected := ClusterSettings{
		LogIndexPrefix:            "test-prefix",
		LogStorageLifetime:        3333,
		MonitoringStorageLifetime: 2222,
	}
	c, err := SecretToClusterSettings(cnf1)
	assert.Nil(t, err)
	assert.Equal(t, expected, c)

	cnf2 := apiv1.Secret{
		Data: map[string][]byte{
			"log-index-prefix":            []byte("test-prefix"),
			"log-storage-lifetime":        []byte("err-data"),
			"monitoring-storage-lifetime": []byte("2222"),
		},
	}
	c, err = SecretToClusterSettings(cnf2)
	assert.NotNil(t, err)
}

func TestGetClusterSettings(t *testing.T) {
	expected := ClusterSettings{
		MonitoringStorageLifetime: 2222,
		LogStorageLifetime:        3333,
		LogIndexPrefix:            "test-",
	}

	s := &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysecret",
			Namespace: "kube-system",
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"username":                []byte("username"),
			"password":                []byte("password"),
			LogIndexPrefix:            []byte(expected.LogIndexPrefix),
			LogStorageLifetime:        []byte(fmt.Sprintf("%v", expected.LogStorageLifetime)),
			MonitoringStorageLifetime: []byte(fmt.Sprintf("%v", expected.MonitoringStorageLifetime)),
		},
	}
	cs, err := getClusterSettings(fake.NewSimpleClientset(s), s.Name, s.Namespace)
	assert.Nil(t, err)
	assert.Equal(t, expected, cs)

	_, err = getClusterSettings(fake.NewSimpleClientset(s), "notpresent", s.Namespace)
	assert.NotNil(t, err)
}
