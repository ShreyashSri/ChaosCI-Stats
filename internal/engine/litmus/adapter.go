package litmus

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ShreyashSri/ChaosCI-Stats/internal/engine"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/store"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"sigs.k8s.io/yaml"
)

type Adapter struct {
	client dynamic.Interface
}

func NewAdapter(client dynamic.Interface) *Adapter {
	return &Adapter{
		client: client,
	}
}

func (a *Adapter) Apply(ctx context.Context, exp store.Experiment) error {
	obj, gvr, err := a.parseYaml(exp)
	if err != nil {
		return err
	}

	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["chaosci.io/run-id"] = exp.RunID.String
	labels["chaosci.io/experiment-id"] = fmt.Sprintf("%d", exp.ID)
	obj.SetLabels(labels)

	_, err = a.client.Resource(*gvr).Namespace(namespace).Create(ctx, obj, v1.CreateOptions{})
	return err
}

func (a *Adapter) Watch(ctx context.Context, exp store.Experiment) (<-chan engine.Result, error) {
	ch := make(chan engine.Result)
	obj, gvr, err := a.parseYaml(exp)
	if err != nil {
		return nil, err
	}

	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}
	name := obj.GetName()

	go func() {
		defer close(ch)
		watcher, err := a.client.Resource(*gvr).Namespace(namespace).Watch(ctx, v1.ListOptions{
			FieldSelector: "metadata.name=" + name,
		})
		if err != nil {
			ch <- engine.Result{
				ExperimentID: exp.ID,
				Status:       "error",
				Message:      fmt.Sprintf("failed to watch Litmus experiment: %v", err),
			}
			return
		}
		defer watcher.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.ResultChan():
				if !ok {
					return
				}
				u, ok := event.Object.(*unstructured.Unstructured)
				if !ok {
					continue
				}

				statusPhase, found, _ := unstructured.NestedString(u.Object, "status", "engineStatus")
				if !found || statusPhase == "" {
					continue
				}

				status := statusPhase
				if statusPhase == "completed" {
					status = "success"
				} else if statusPhase == "stopped" {
					status = "failed"
				}

				ch <- engine.Result{
					ExperimentID: exp.ID,
					Status:       status,
					Message:      fmt.Sprintf("Litmus status updated to %s", statusPhase),
				}

				if statusPhase == "completed" || statusPhase == "stopped" || statusPhase == "error" {
					return
				}
			}
		}
	}()

	return ch, nil
}

func (a *Adapter) Cleanup(ctx context.Context, exp store.Experiment) error {
	obj, gvr, err := a.parseYaml(exp)
	if err != nil {
		return err
	}

	namespace := obj.GetNamespace()
	if namespace == "" {
		namespace = "default"
	}

	return a.client.Resource(*gvr).Namespace(namespace).Delete(ctx, obj.GetName(), v1.DeleteOptions{})
}

func (a *Adapter) parseYaml(exp store.Experiment) (*unstructured.Unstructured, *schema.GroupVersionResource, error) {
	file := "chaos.yaml"
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read experiment file: %w", err)
	}

	var obj unstructured.Unstructured
	if err := yaml.Unmarshal(data, &obj.Object); err != nil {
		return nil, nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	gvk := obj.GroupVersionKind()
	gvr := schema.GroupVersionResource{
		Group:    gvk.Group,
		Version:  gvk.Version,
		Resource: strings.ToLower(gvk.Kind) + "s",
	}

	return &obj, &gvr, nil
}
