package slackbot

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"

	jenkinsv1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	"github.com/jenkins-x/jx/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (c *GlobalClients) getPipelineActivities(org string, repo string, prn int) (*jenkinsv1.PipelineActivityList, error) {
	return c.JXClient.JenkinsV1().PipelineActivities(c.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("owner=%s, branch=PR-%d, repository=%s", org, prn, repo),
	})
}

// WatchActivities watches for pipeline activities
func (o *SlackBotOptions) WatchActivities() chan struct{} {

	jxClient := o.JXClient

	log.Logger().Infof("Watching pipelines in namespace %s\n", o.Namespace)

	activity := &jenkinsv1.PipelineActivity{}
	_, controller := cache.NewInformer(
		cache.NewListWatchFromClient(jxClient.JenkinsV1().RESTClient(), "pipelineactivities", o.Namespace,
			fields.Everything()),
		activity,
		time.Minute*10,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				o.onObj(obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				o.onObj(newObj)
			},
			DeleteFunc: func(obj interface{}) {
			},
		},
	)

	stop := make(chan struct{})
	go controller.Run(stop)

	return stop
}

func (o *SlackBotOptions) onObj(obj interface{}) {
	activity, ok := obj.(*jenkinsv1.PipelineActivity)
	if !ok {
		log.Logger().Infof("Object is not a PipelineActivity %#v\n", obj)
		return
	}
	err := o.PipelineMessage(activity)
	if err != nil {
		log.Logger().Warnf("%v\n", err)
	}
	err = o.ReviewRequestMessage(activity)
	if err != nil {
		log.Logger().Warnf("%v\n", err)
	}
}
