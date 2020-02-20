package test

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MockClient struct {
	updateFns map[runtime.Object]func() error
}

func NewMockClient() MockClient {
	return MockClient{
		updateFns: map[runtime.Object]func() error{},
	}
}

func (m *MockClient) Get(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
	return nil
}

func (m *MockClient) List(ctx context.Context, list runtime.Object, opts ...client.ListOption) error {
	return nil
}

func (m *MockClient) Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error {
	return nil
}

func (m *MockClient) Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error {
	return nil
}

func (m *MockClient) Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error {
	return m.updateFns[obj]()
}

func (m *MockClient) WhenUpdate(obj runtime.Object, then func() error) {
	m.updateFns[obj] = then
}

func (m *MockClient) Patch(ctx context.Context, obj runtime.Object, patch client.Patch, opts ...client.PatchOption) error {
	return nil
}

func (m *MockClient) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...client.DeleteAllOfOption) error {
	return nil
}

func (m *MockClient) Status() client.StatusWriter {
	return nil
}
