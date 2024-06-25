package operator

import (
	"context"
	log "github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	nacosgroupv1alpha1 "nacos.io/nacos-operator/api/v1alpha1"
	myErrors "nacos.io/nacos-operator/pkg/errors"
	"nacos.io/nacos-operator/pkg/service/k8s"
	"nacos.io/nacos-operator/pkg/util/contains"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IOperatorClient interface {
	IKindClient
	ICheckClient
	IHealClient
	IStatusClient
	IUpdateClient
}

type OperatorClient struct {
	KindClient   *KindClient
	CheckClient  *CheckClient
	HealClient   *HealClient
	StatusClient *StatusClient
	UpdateClient *UpdateClient
}

func NewOperatorClient(logger log.Logger, clientset *kubernetes.Clientset, s *runtime.Scheme, client client.Client) *OperatorClient {
	service := k8s.NewK8sService(clientset, logger)
	return &OperatorClient{
		// 资源客户端
		KindClient: NewKindClient(logger, service, s),
		// 检测客户端
		CheckClient: NewCheckClient(logger, service),
		// 状态客户端
		StatusClient: NewStatusClient(logger, service, client),
		// 维护客户端
		HealClient: NewHealClient(logger, service),
		// 更新非状态客户端
		UpdateClient: NewUpdateClient(logger, service, client),
	}
}

func (c *OperatorClient) MakeEnsure(nacos *nacosgroupv1alpha1.Nacos) {
	// 验证CR字段
	c.KindClient.ValidationField(nacos)

	switch nacos.Spec.Type {
	case TYPE_STAND_ALONE:
		c.KindClient.EnsureConfigmap(nacos)
		c.KindClient.EnsureStatefulset(nacos)
		c.KindClient.EnsureService(nacos)
		if nacos.Spec.Database.TypeDatabase == "mysql" && nacos.Spec.MysqlInitImage != "" {
			c.KindClient.EnsureMysqlConfigMap(nacos)
			c.KindClient.EnsureJob(nacos)
		}
	case TYPE_CLUSTER:
		c.KindClient.EnsureConfigmap(nacos)
		c.KindClient.EnsureStatefulsetCluster(nacos)
		c.KindClient.EnsureHeadlessServiceCluster(nacos)
		c.KindClient.EnsureClientService(nacos)
		if nacos.Spec.Database.TypeDatabase == "mysql" && nacos.Spec.MysqlInitImage != "" {
			c.KindClient.EnsureMysqlConfigMap(nacos)
			c.KindClient.EnsureJob(nacos)
		}
	default:
		panic(myErrors.New(myErrors.CODE_PARAMETER_ERROR, myErrors.MSG_PARAMETER_ERROT, "nacos.Spec.Type", nacos.Spec.Type))
	}
}

func (c *OperatorClient) HandlerFinalizers(nacos *nacosgroupv1alpha1.Nacos) {
	finalizer := "cleanUpNacosPvc"
	if nacos.Spec.Database.TypeDatabase != "embedded" {
		return
	}
	if nacos.DeletionTimestamp.IsZero() {
		if !contains.ContainString(nacos.ObjectMeta.Finalizers,finalizer){
			nacos.ObjectMeta.Finalizers = append(nacos.ObjectMeta.Finalizers,finalizer)
		   c.UpdateClient.Update(nacos)
		}
	} else {
		if contains.ContainString(nacos.ObjectMeta.Finalizers,finalizer) {
			c.CleanAllPvcs(nacos)
			nacos.ObjectMeta.Finalizers = contains.RemoveString(nacos.ObjectMeta.Finalizers, finalizer)
			c.UpdateClient.Update(nacos)
		}
	}
}
func (c *OperatorClient) getPvcList(nacos *nacosgroupv1alpha1.Nacos)(Pvcs corev1.PersistentVolumeClaimList, err error) {
	selector, err := metav1.LabelSelectorAsSelector(&metav1.LabelSelector{
		MatchLabels: map[string]string{"app": nacos.GetName(),"component": "nacos"},
	})
	pvcListOps := &client.ListOptions{
		Namespace: nacos.Namespace,
		LabelSelector: selector,
	}
	pvcList := &corev1.PersistentVolumeClaimList{}
	err = c.UpdateClient.client.List(context.TODO(),pvcList,pvcListOps)
	return *pvcList, err
}
func (c *OperatorClient) deletePvc(pvcItem corev1.PersistentVolumeClaim){
	pvcDelete := &corev1.PersistentVolumeClaim{
		ObjectMeta : metav1.ObjectMeta{
			Name: pvcItem.Name,
			Namespace: pvcItem.Namespace,
		},
	}
	if err := c.UpdateClient.client.Delete(context.TODO(), pvcDelete);err != nil{
		c.UpdateClient.logger.V(0).Error(err,"delete pvc error")
	}
}

func (c *OperatorClient) CleanAllPvcs(nacos *nacosgroupv1alpha1.Nacos) {
	pvcs, err := c.getPvcList(nacos)
	c.UpdateClient.logger.V(0).Info("get pvc ","pvs",pvcs)
	if err != nil {
		myErrors.EnsureNormalMyError(err, myErrors.CODE_CLUSTER_FAILE)
	}
	for _, pvcItem := range pvcs.Items {
		c.deletePvc(pvcItem)
	}
}


func (c *OperatorClient) PreCheck(nacos *nacosgroupv1alpha1.Nacos) {
	switch nacos.Status.Phase {
	case nacosgroupv1alpha1.PhaseFailed:
		// 失败，需要修复
		c.HealClient.MakeHeal(nacos)
	case nacosgroupv1alpha1.PhaseNone:
		// 初始化
		nacos.Status.Phase = nacosgroupv1alpha1.PhaseCreating
		panic(myErrors.New(myErrors.CODE_NORMAL, ""))
	case nacosgroupv1alpha1.PhaseScale:
	default:
		// TODO
	}
}

func (c *OperatorClient) CheckAndMakeHeal(nacos *nacosgroupv1alpha1.Nacos) {
	// 检查kind
	pods := c.CheckClient.CheckKind(nacos)
	// 检查nacos
	c.CheckClient.CheckNacos(nacos, pods)
}

func (c *OperatorClient) UpdateStatus(nacos *nacosgroupv1alpha1.Nacos) {
	c.StatusClient.UpdateStatusRunning(nacos)
}
