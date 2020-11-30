package utils

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

func CreateConfigMap(opt TestOptions, isHub bool, cm *corev1.ConfigMap) error {
	clientKube := getKubeClient(opt, isHub)
	_, err := clientKube.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(cm)
	return err
}

func GetConfigMap(opt TestOptions, isHub bool, name string,
	namespace string) (error, *corev1.ConfigMap) {
	clientKube := getKubeClient(opt, isHub)
	cm, err := clientKube.CoreV1().ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		klog.Errorf("Failed to get configmap %s in namespace %s due to %v", name, namespace, err)
	}
	return err, cm
}

func DeleteConfigMap(opt TestOptions, isHub bool, name string, namespace string) error {
	clientKube := getKubeClient(opt, isHub)
	err := clientKube.CoreV1().ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		klog.Errorf("Failed to delete configmap %s in namespace %s due to %v", name, namespace, err)
	}
	return err
}
