package slackbot

import (
	"fmt"

	jenkinsv1 "github.com/jenkins-x/jx/pkg/apis/jenkins.io/v1"
	informers "github.com/jenkins-x/jx/pkg/client/informers/externalversions"
	"github.com/jenkins-x/jx/pkg/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (c *GlobalClients) getPipelineActivities(org string, repo string, prn int) (*jenkinsv1.PipelineActivityList, error) {
	return c.JXClient.JenkinsV1().PipelineActivities(c.Namespace).List(metav1.ListOptions{
		LabelSelector: fmt.Sprintf("owner=%s, branch=PR-%d, repository=%s", org, prn, repo),
	})
}

// WatchActivities watches for pipeline activities
func (o *SlackBotOptions) WatchActivities() chan struct{} {

	log.Logger().Infof("Watching pipeline activities in namespace %s and slackbot config %s", o.Namespace, o.Name)

	factory := informers.NewSharedInformerFactoryWithOptions(o.JXClient, 0, informers.WithNamespace(o.Namespace))

	informer := factory.Jenkins().V1().PipelineActivities().Informer()

	stopper := make(chan struct{})

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    o.onObj,
		UpdateFunc: o.onUpdate,
	})

	go informer.Run(stopper)

	return stopper
}

func (o *SlackBotOptions) onObj(obj interface{}) {

	activity, ok := obj.(*jenkinsv1.PipelineActivity)
	if !ok {
		log.Logger().Infof("Object is not a PipelineActivity %#v\n", obj)
		return
	}
	log.Logger().Infof("activity %s ", activity.Name)
	err := o.PipelineMessage(activity)
	if err != nil {
		log.Logger().Warnf("%v\n", err)
	}
	err = o.ReviewRequestMessage(activity)
	if err != nil {
		log.Logger().Warnf("%v\n", err)
	}
}

func (o SlackBotOptions) onUpdate(_ interface{}, newObj interface{}) {
	o.onObj(newObj)
}
