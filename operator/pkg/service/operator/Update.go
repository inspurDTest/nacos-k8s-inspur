package operator

import (
	"context"
	"sigs.k8s.io/controller-runtime/pkg/client"

	log "github.com/go-logr/logr"
	nacosgroupv1alpha1 "nacos.io/nacos-operator/api/v1alpha1"
	myErrors "nacos.io/nacos-operator/pkg/errors"
	"nacos.io/nacos-operator/pkg/service/k8s"
)

type IUpdateClient interface {
}

type UpdateClient struct {
	logger log.Logger
	client client.Client
}

func NewUpdateClient(logger log.Logger, k8sService k8s.Services, client client.Client) *UpdateClient {
	return &UpdateClient{
		client: client,
		logger: logger,
	}
}

// 更新nacos非状态信息
func (c *UpdateClient) Update(nacos *nacosgroupv1alpha1.Nacos) {
	UpdateLastEvent(nacos, 200, "", true)
	// TODO
	myErrors.EnsureNormal(c.client.Update(context.TODO(), nacos))
}

