package configmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type configMapGetter struct {
	cm corev1.ConfigMap
}

func (c configMapGetter) GetConfigMap(objectKey client.ObjectKey) (corev1.ConfigMap, error) {
	if c.cm.Name == objectKey.Name && c.cm.Namespace == objectKey.Namespace {
		return c.cm, nil
	}
	return corev1.ConfigMap{}, notFoundError()
}

func newGetter(cm corev1.ConfigMap) Getter {
	return configMapGetter{
		cm: cm,
	}
}

func TestReadKey(t *testing.T) {
	getter := newGetter(
		Builder().
			SetName("name").
			SetNamespace("namespace").
			SetField("key1", "value1").
			SetField("key2", "value2").
			Build(),
	)

	value, err := ReadKey(getter, "key1", nsName("namespace", "name"))
	assert.Equal(t, "value1", value)
	assert.NoError(t, err)

	value, err = ReadKey(getter, "key2", nsName("namespace", "name"))
	assert.Equal(t, "value2", value)
	assert.NoError(t, err)

	_, err = ReadKey(getter, "key3", nsName("namespace", "name"))
	assert.Error(t, err)
}

func TestReadData(t *testing.T) {
	getter := newGetter(
		Builder().
			SetName("name").
			SetNamespace("namespace").
			SetField("key1", "value1").
			SetField("key2", "value2").
			Build(),
	)

	data, err := ReadData(getter, nsName("namespace", "name"))
	assert.NoError(t, err)

	assert.Contains(t, data, "key1")
	assert.Contains(t, data, "key2")

	assert.Equal(t, "value1", data["key1"])
	assert.Equal(t, "value2", data["key2"])
}

type configMapGetUpdater struct {
	cm corev1.ConfigMap
}

func (c configMapGetUpdater) GetConfigMap(objectKey client.ObjectKey) (corev1.ConfigMap, error) {
	if c.cm.Name == objectKey.Name && c.cm.Namespace == objectKey.Namespace {
		return c.cm, nil
	}
	return corev1.ConfigMap{}, notFoundError()
}

func (c configMapGetUpdater) UpdateConfigMap(cm corev1.ConfigMap) error {
	c.cm = cm
	return nil
}

func newGetUpdater(cm corev1.ConfigMap) GetUpdater {
	return configMapGetUpdater{
		cm: cm,
	}
}

func TestUpdateField(t *testing.T) {
	getUpdater := newGetUpdater(
		Builder().
			SetName("name").
			SetNamespace("namespace").
			SetField("field1", "value1").
			SetField("field2", "value2").
			Build(),
	)
	err := UpdateField(getUpdater, nsName("namespace", "name"), "field1", "newValue")
	assert.NoError(t, err)
	val, _ := ReadKey(getUpdater, "field1", nsName("namespace", "name"))
	assert.Equal(t, "newValue", val)
	val2, _ := ReadKey(getUpdater, "field2", nsName("namespace", "name"))
	assert.Equal(t, "value2", val2)
}

func nsName(namespace, name string) types.NamespacedName {
	return types.NamespacedName{Name: name, Namespace: namespace}
}

func notFoundError() error {
	return &errors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
}
