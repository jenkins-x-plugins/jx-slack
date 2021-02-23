package slackbot

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	jenkinsv1 "github.com/jenkins-x/jx-api/v4/pkg/apis/jenkins.io/v1"
	informers "github.com/jenkins-x/jx-api/v4/pkg/client/informers/externalversions"
	"github.com/jenkins-x/jx-logging/v3/pkg/log"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (o *Options) getPipelineActivities(ctx context.Context, org string, repo string, prn int) (*jenkinsv1.PipelineActivityList, error) {
	return o.JXClient.JenkinsV1().PipelineActivities(o.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("owner=%s, branch=PR-%d, repository=%s", org, prn, repo),
	})
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

	//go informer.Run(stopper)
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

func (o Options) onUpdate(_ interface{}, newObj interface{}) {
	o.onObj(newObj)
}
