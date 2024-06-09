package app_orchestrator

import (
	"context"
	"fmt"
	"github.com/jmhbh/argo-app-orchestrator/app_orchestrator/utils"
	"github.com/jmhbh/argo-app-orchestrator/k8sclient"
	. "github.com/jmhbh/argo-app-orchestrator/types"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"os"
	"path"
	"slices"
)

var argoAppSetGVR = schema.GroupVersionResource{
	Group:    "argoproj.io",
	Version:  "v1alpha1",
	Resource: "applicationsets",
}

var argoCdNs = "argocd"

type Reconciler struct {
	clientSet        *kubernetes.Clientset
	dynamicClient    *dynamic.DynamicClient
	userMetadataChan chan UserMetadata
	kickChan         chan struct{}
	users            []string // TODO: store in durable storage instead of in memory
}

func Start(ctx context.Context, params Params) error {
	logger := ctx.Value(LoggerKey{}).(*zap.SugaredLogger)
	// must run in a kube cluster otherwise will panic
	clientSet, dynamicClient, err := k8sclient.InitClients()
	if err != nil {
		logger.Errorf("failed to init k8clients in orchestrator: %s", err.Error())
		return err
	}

	rec := NewReconciler(clientSet, dynamicClient, params)
	return rec.Run(ctx)
}

func NewReconciler(clientSet *kubernetes.Clientset, dynamicClient *dynamic.DynamicClient, params Params) *Reconciler {
	return &Reconciler{
		clientSet:        clientSet,
		dynamicClient:    dynamicClient,
		userMetadataChan: params.UserMetadataChan,
		kickChan:         params.KickChan,
	}
}

func (r *Reconciler) Run(ctx context.Context) error {
	logger := ctx.Value(LoggerKey{}).(*zap.SugaredLogger)
	for {
		select {
		case <-ctx.Done():
			return nil
		case userMetadata := <-r.userMetadataChan:
			if err := r.applyArgoAppSet(ctx, userMetadata); err != nil {
				logger.Errorf("encountered error applying argo app set: %s", err.Error())
			}
		}
	}
}

func (r *Reconciler) applyArgoAppSet(ctx context.Context, userMetadata UserMetadata) error {
	// check if user exists in the cache if not add it to the list and
	// create our new appset yaml then apply it
	newUser := userMetadata.Name
	if slices.Contains(r.users, newUser) {
		return fmt.Errorf("user %s already exists", newUser)
	} else {
		r.users = append(r.users, newUser)
	}

	// load yaml and template it with all users in our cache
	// the appsets list generator will contain all users in our cache and create deployments
	// for each user
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	appSetYaml, err := os.ReadFile(path.Join(cwd, "/app_orchestrator/applicationset.yaml"))
	if err != nil {
		return err
	}

	buffer, err := utils.TemplateYaml(r.users, appSetYaml)
	if err != nil {
		return err
	}

	obj := &unstructured.Unstructured{}
	err = yaml.NewYAMLOrJSONDecoder(buffer, buffer.Len()).Decode(obj)
	if err != nil {
		return err
	}
	_, err = r.dynamicClient.
		Resource(argoAppSetGVR).
		Namespace(argoCdNs).
		Apply(context.Background(), "supermario-app-set", obj, v1.ApplyOptions{FieldManager: "argo-app-orchestrator"})

	if err != nil {
		return err
	}

	// TODO: right now we just kick the channel to tell the webserver we modified the argo appset in the future we can watch the replicaset wait until its ready and send the pod info back which can be used in the response
	r.kickChan <- struct{}{}
	return nil
}
