package cnb

import (
	"context"
	"fmt"
	"io"
	"testing"

	registryv1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	k8scorev1 "k8s.io/api/core/v1"

	buildapi "github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
)

type fakeLayer struct {
	digest string
	diffID string
	size   int64
}

func (f fakeLayer) Digest() (registryv1.Hash, error) {
	return registryv1.NewHash(f.digest)
}

func (f fakeLayer) DiffID() (registryv1.Hash, error) {
	return registryv1.NewHash(f.diffID)
}

func (f fakeLayer) Size() (int64, error) {
	return f.size, nil
}

func (f fakeLayer) MediaType() (types.MediaType, error) {
	return types.DockerLayer, nil
}

func (f fakeLayer) Compressed() (io.ReadCloser, error) {
	panic("Not implemented For Tests")
}

func (f fakeLayer) Uncompressed() (io.ReadCloser, error) {
	panic("Not implemented For Tests")
}

type buildpackRefContainer struct {
	Ref       buildapi.BuilderBuildpackRef
	Buildpack K8sRemoteBuildpack
}

type fakeResolver struct {
	buildpacks         map[string]K8sRemoteBuildpack
	extensions         map[string]K8sRemoteBuildpack
	observedGeneration int64
}

func (r *fakeResolver) resolveBuildpack(ref buildapi.BuilderBuildpackRef) (K8sRemoteBuildpack, error) {
	buildpack, ok := r.buildpacks[fmt.Sprintf("%s@%s", ref.Id, ref.Version)]
	if !ok {
		return K8sRemoteBuildpack{}, errors.New("buildpack not found")
	}
	return buildpack, nil
}

func (r *fakeResolver) resolveExtension(ref buildapi.BuilderBuildpackRef) (K8sRemoteBuildpack, error) {
	extension, ok := r.extensions[fmt.Sprintf("%s@%s", ref.Id, ref.Version)]
	if !ok {
		return K8sRemoteBuildpack{}, errors.New("extension not found")
	}
	return extension, nil
}

func (f *fakeResolver) AddBuildpack(t *testing.T, ref buildapi.BuilderBuildpackRef, buildpack K8sRemoteBuildpack) {
	t.Helper()
	assert.NotEqual(t, ref.Id, "", "buildpack ref missing id")
	f.buildpacks[fmt.Sprintf("%s@%s", ref.Id, ref.Version)] = buildpack
}

func (f *fakeResolver) AddExtension(t *testing.T, ref buildapi.BuilderBuildpackRef, extension K8sRemoteBuildpack) {
	t.Helper()
	assert.NotEqual(t, ref.Id, "", "extension ref missing id")
	f.extensions[fmt.Sprintf("%s@%s", ref.Id, ref.Version)] = extension
}

func (r *fakeResolver) ClusterStoreObservedGeneration() int64 {
	return r.observedGeneration
}

func makeRef(id, version string) buildapi.BuilderBuildpackRef {
	return buildapi.BuilderBuildpackRef{
		BuildpackRef: corev1alpha1.BuildpackRef{
			BuildpackInfo: corev1alpha1.BuildpackInfo{
				Id:      id,
				Version: version,
			},
		},
	}
}

func makeObjectRef(name, kind, id, version string) buildapi.BuilderBuildpackRef {
	return buildapi.BuilderBuildpackRef{
		ObjectReference: k8scorev1.ObjectReference{
			Name: name,
			Kind: kind,
		},
		BuildpackRef: corev1alpha1.BuildpackRef{
			BuildpackInfo: corev1alpha1.BuildpackInfo{
				Id:      id,
				Version: version,
			},
		},
	}
}

type fakeFetcher struct {
	buildpacks         map[string][]buildpackLayer
	extensions         map[string][]buildpackLayer
	observedGeneration int64
}

func (f *fakeFetcher) ResolveAndFetchBuildpack(_ context.Context, bp buildapi.BuilderBuildpackRef) (RemoteBuildpackInfo, error) {
	bpLayers, ok := f.buildpacks[fmt.Sprintf("%s@%s", bp.Id, bp.Version)]
	if ok {
		return RemoteBuildpackInfo{
			BuildpackInfo: buildpackInfoInLayers(bpLayers, bp.Id, bp.Version),
			Layers:        bpLayers,
		}, nil
	}
	return RemoteBuildpackInfo{}, errors.New("buildpack not found")
}

func (f *fakeFetcher) ResolveAndFetchExtension(_ context.Context, ext buildapi.BuilderBuildpackRef) (RemoteBuildpackInfo, error) {
	extLayers, ok := f.extensions[fmt.Sprintf("%s@%s", ext.Id, ext.Version)]
	if ok {
		return RemoteBuildpackInfo{
			BuildpackInfo: buildpackInfoInLayers(extLayers, ext.Id, ext.Version),
			Layers:        extLayers,
		}, nil
	}
	return RemoteBuildpackInfo{}, errors.New("extension not found")
}

func (f *fakeFetcher) ClusterStoreObservedGeneration() int64 {
	return f.observedGeneration
}

func (f *fakeFetcher) UsedObjects() []k8scorev1.ObjectReference {
	return nil
}

func (f *fakeFetcher) resolveBuildpack(ref buildapi.BuilderBuildpackRef) (K8sRemoteBuildpack, error) {
	panic("Not implemented For Tests")
}

func (f *fakeFetcher) resolveExtension(ref buildapi.BuilderBuildpackRef) (K8sRemoteBuildpack, error) {
	panic("Not implemented For Tests")
}

func (f *fakeFetcher) AddBuildpack(t *testing.T, id, version string, layers []buildpackLayer) {
	t.Helper()
	f.buildpacks[fmt.Sprintf("%s@%s", id, version)] = layers
}

func (f *fakeFetcher) AddExtension(t *testing.T, id, version string, layers []buildpackLayer) {
	t.Helper()
	f.extensions[fmt.Sprintf("%s@%s", id, version)] = layers
}
