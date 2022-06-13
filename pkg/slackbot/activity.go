package slackbot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jenkins-x/jx-helpers/v3/pkg/kube/activities"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	informers "github.com/jenkins-x/jx-api/v4/pkg/client/informers/externalversions"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (o *Options) getPipelineActivities(ctx context.Context, owner, repo string, prn int) (*jenkinsv1.PipelineActivityList, error) {
	branch := fmt.Sprintf("PR-%d", prn)

	list, err := o.JXClient.JenkinsV1().PipelineActivities(o.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return list, err
	}

	if list != nil {
		results := &jenkinsv1.PipelineActivityList{
			TypeMeta: list.TypeMeta,
			ListMeta: list.ListMeta,
		}

		for i := range list.Items {
			r := &list.Items[i]
			// lets default the properties if missing from the labels
			activities.DefaultValues(r)
			if owner == r.Spec.GitOwner && repo == r.Spec.GitRepository && branch == r.Spec.GitBranch {
				results.Items = append(results.Items, *r)
			}
		}
		return results, nil
	}
	return nil, nil
}

func (o *Options) previousPipelineFailed(activity *jenkinsv1.PipelineActivity) (bool, error) {
	build := activity.Spec.Build
	if build == "" || build == "1" {
		return false, nil
	}
	buildNumber, err := strconv.Atoi(build)
	if err != nil {
		return false, nil
	}
	if buildNumber <= 1 {
		return false, nil
	}

	// lets use the previous build number
	name := activity.Name
	idx := strings.LastIndex(name, "-")
	if idx <= 0 {
		return false, nil
	}
	previousName := name[0:idx] + "-" + strconv.Itoa(buildNumber-1)

	ctx := context.TODO()
	previous, err := o.JXClient.JenkinsV1().PipelineActivities(o.Namespace).Get(ctx, previousName, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, errors.Wrapf(err, "failed to find PipelineActivity %s in namespace %s", previousName, o.Namespace)
	}
	if previous == nil {
		return false, nil
	}
	if previous.Spec.Status == jenkinsv1.ActivityStatusTypeFailed || previous.Spec.Status == jenkinsv1.ActivityStatusTypeError {
		// TODO should we record the previous message so we can update it?
		return true, nil
	}
	return false, nil
}

// WatchActivities watches for pipeline activities
func (o *Options) WatchActivities() chan struct{} {
	log.Logger().Infof("Watching pipeline activities in namespace %s and slackbot config %s", o.Namespace, o.Name)

	factory := informers.NewSharedInformerFactoryWithOptions(o.JXClient, 0, informers.WithNamespace(o.Namespace))

	informer := factory.Jenkins().V1().PipelineActivities().Informer()

	stopper := make(chan struct{})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.onObj,
		UpdateFunc: o.onUpdate,
	})

	informer.Run(stopper)
	return stopper
}

func (o *Options) onObj(obj interface{}) {
	activity, ok := obj.(*jenkinsv1.PipelineActivity)
	if !ok {
		log.Logger().Infof("Object is not a PipelineActivity %#v\n", obj)
		return
	}
	log.Logger().Debugf("activity %s ", activity.Name)
	err := o.PipelineMessage(activity)
	if err != nil {
		log.Logger().Warnf("%v\n", err)
	}
	err = o.ReviewRequestMessage(activity)
	if err != nil {
		log.Logger().Warnf("%v\n", err)
	}
}

func (o *Options) onUpdate(_, newObj interface{}) {
	o.onObj(newObj)
}
