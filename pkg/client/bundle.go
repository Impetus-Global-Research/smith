package client

import (
	"time"

	smith_v1 "github.com/atlassian/smith/pkg/apis/smith/v1"
	smithClient_v1 "github.com/atlassian/smith/pkg/client/clientset_generated/clientset/typed/smith/v1"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

func BundleInformer(bundleClient smithClient_v1.BundlesGetter, namespace string, resyncPeriod time.Duration) cache.SharedIndexInformer {
	bundlesApi := bundleClient.Bundles(namespace)
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options meta_v1.ListOptions) (runtime.Object, error) {
				return bundlesApi.List(options)
			},
			WatchFunc: func(options meta_v1.ListOptions) (watch.Interface, error) {
				return bundlesApi.Watch(options)
			},
		},
		&smith_v1.Bundle{},
		resyncPeriod,
		cache.Indexers{})
}
