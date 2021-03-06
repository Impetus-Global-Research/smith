package controller_test

import (
	"context"
	"testing"

	smith_v1 "github.com/atlassian/smith/pkg/apis/smith/v1"
	"github.com/atlassian/smith/pkg/controller"
	sc_v1b1 "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kube_testing "k8s.io/client-go/testing"
	"k8s.io/client-go/tools/cache"
)

// Should not perform any creates/updates/deletes on resources after bundle
// is marked with deletionTimestamp and "foregroundDeletion" finalizer
func TestNoActionsForResourcesWhenForegroundDeletion(t *testing.T) {
	t.Parallel()
	now := meta_v1.Now()
	tc := testCase{
		mainClientObjects: []runtime.Object{
			configMapNeedsDelete(),
			configMapNeedsUpdate(),
		},
		scClientObjects: []runtime.Object{
			serviceInstance(false, false, true),
		},
		bundle: &smith_v1.Bundle{
			ObjectMeta: meta_v1.ObjectMeta{
				Name:              bundle1,
				Namespace:         testNamespace,
				UID:               bundle1uid,
				DeletionTimestamp: &now,
				Finalizers:        []string{meta_v1.FinalizerDeleteDependents},
			},
			Spec: smith_v1.BundleSpec{
				Resources: []smith_v1.Resource{
					{
						Name: resSi1,
						Spec: smith_v1.ResourceSpec{
							Object: &sc_v1b1.ServiceInstance{
								TypeMeta: meta_v1.TypeMeta{
									Kind:       "ServiceInstance",
									APIVersion: sc_v1b1.SchemeGroupVersion.String(),
								},
								ObjectMeta: meta_v1.ObjectMeta{
									Name: si1,
								},
								Spec: serviceInstanceSpec,
							},
						},
					},
					{
						Name:      resMapNeedsAnUpdate,
						DependsOn: []smith_v1.ResourceName{resSi1},
						Spec: smith_v1.ResourceSpec{
							Object: &core_v1.ConfigMap{
								TypeMeta: meta_v1.TypeMeta{
									Kind:       "ConfigMap",
									APIVersion: core_v1.SchemeGroupVersion.String(),
								},
								ObjectMeta: meta_v1.ObjectMeta{
									Name: mapNeedsAnUpdate,
								},
							},
						},
					},
				},
			},
		},
		namespace:            testNamespace,
		enableServiceCatalog: false,
		test: func(t *testing.T, ctx context.Context, cntrlr *controller.BundleController, tc *testCase, prepare func(ctx context.Context)) {
			prepare(ctx)
			key, err := cache.MetaNamespaceKeyFunc(tc.bundle)
			require.NoError(t, err)
			_, err = cntrlr.ProcessKey(tc.logger, key)
			assert.NoError(t, err)

			actions := tc.bundleFake.Actions()
			require.Len(t, actions, 2)
			assert.Implements(t, (*kube_testing.ListAction)(nil), actions[0])
			assert.Implements(t, (*kube_testing.WatchAction)(nil), actions[1])
		},
	}
	tc.run(t)
}
