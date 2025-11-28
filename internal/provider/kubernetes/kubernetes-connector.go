package kubernetes

import (
	"context"
	"fmt"

	"github.com/jaku01/caddyservicediscovery/internal/caddy"
	"github.com/jaku01/caddyservicediscovery/internal/provider"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Connector struct {
	ClientSet *kubernetes.Clientset
	ctx       context.Context
}

type ServiceEvent struct {
	Type    watch.EventType
	Service *corev1.Service
}

func NewKubernetesConnector() (*Connector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &Connector{
		ClientSet: clientSet,
		ctx:       context.Background(),
	}, nil
}

func (c *Connector) GetRoutes() ([]caddy.Route, error) {
	services, err := c.ClientSet.CoreV1().Services(metav1.NamespaceAll).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	routes := make([]caddy.Route, 0, len(services.Items))
	for _, svc := range services.Items {
		domain := svc.Labels["domain"]
		if domain == "" {
			continue
		}
		if len(svc.Spec.Ports) == 0 {
			continue
		}

		upstream := fmt.Sprintf(
			"%s.%s.svc.cluster.local:%d",
			svc.Name,
			svc.Namespace,
			svc.Spec.Ports[0].Port,
		)

		route := caddy.NewReverseProxyRoute(domain, upstream)
		routes = append(routes, route)
	}

	return routes, nil
}

func (c *Connector) GetEventChannel() <-chan provider.LifecycleEvent {
	lifecycleEvents := make(chan provider.LifecycleEvent)

	go func() {
		defer close(lifecycleEvents)

		watcher, err := c.ClientSet.CoreV1().Services(metav1.NamespaceAll).Watch(c.ctx, metav1.ListOptions{})
		if err != nil {
			return
		}

		defer watcher.Stop()

		for ev := range watcher.ResultChan() {
			select {
			case <-c.ctx.Done():
				return
			default:
			}

			svc, ok := ev.Object.(*corev1.Service)
			if !ok {
				continue
			}

			// map k8s event type to internal LifeCycleEventType
			var eventType provider.EventType
			switch ev.Type {
			case watch.Added:
				eventType = provider.StartEvent
			case watch.Deleted:
				eventType = provider.DieEvent
			default:
				// ignore other event types (Modified etc.)
				continue
			}

			domain := svc.Labels["domain"]
			if domain == "" || len(svc.Spec.Ports) == 0 {
				// nothing to expose -> skip
				continue
			}
			port := int(svc.Spec.Ports[0].Port)

			upstream := fmt.Sprintf(
				"%s.%s.svc.cluster.local:%d",
				svc.Name,
				svc.Namespace,
				svc.Spec.Ports[0].Port,
			)

			lifecycleEvents <- provider.LifecycleEvent{
				ContainerInfo: provider.EndpointInfo{
					Port:     port,
					Domain:   domain,
					Upstream: upstream,
				},
				LifeCycleEventType: eventType,
			}
		}
	}()

	return lifecycleEvents
}
