package contains

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func String(slice []string, s string) bool {
	for _, elem := range slice {
		if elem == s {
			return true
		}
	}
	return false
}

func NamespacedName(nsNames []types.NamespacedName, nsName types.NamespacedName) bool {
	for _, elem := range nsNames {
		if elem == nsName {
			return true
		}
	}
	return false
}

func AccessMode(accessModes []corev1.PersistentVolumeAccessMode, mode corev1.PersistentVolumeAccessMode) bool {
	for _, elem := range accessModes {
		if elem == mode {
			return true
		}
	}
	return false
}
func ContainString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
            return true
        }
	}
	return false
}
func RemoveString(slice []string, str string) (result []string) {
	for _, item := range slice {
		if item == str {
			continue
		}
		result = append(result, item)
	}

}