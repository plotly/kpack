package cnb

import (
	"fmt"
	"sort"

	"github.com/Masterminds/semver/v3"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/pivotal/kpack/pkg/apis/build/v1alpha2"
	corev1alpha1 "github.com/pivotal/kpack/pkg/apis/core/v1alpha1"
	"github.com/pivotal/kpack/pkg/registry"
)

// BuildpackResolver will attempt to resolve a Buildpack reference to a
// Buildpack from either the ClusterStore, Buildpacks, or ClusterBuildpacks
type BuildpackResolver interface {
	resolveBuildpack(ref v1alpha2.BuilderBuildpackRef) (K8sRemoteBuildpack, error)
	resolveExtension(ref v1alpha2.BuilderBuildpackRef) (K8sRemoteBuildpack, error)
	ClusterStoreObservedGeneration() int64
}

type buildpackResolver struct {
	clusterStore      *v1alpha2.ClusterStore
	buildpacks        []*v1alpha2.Buildpack
	clusterBuildpacks []*v1alpha2.ClusterBuildpack
	extensions        []*v1alpha2.Extension
	clusterExtensions []*v1alpha2.ClusterExtension
}

func NewBuildpackResolver(
	clusterStore *v1alpha2.ClusterStore,
	buildpacks []*v1alpha2.Buildpack,
	clusterBuildpacks []*v1alpha2.ClusterBuildpack,
	extensions []*v1alpha2.Extension,
	clusterExtensions []*v1alpha2.ClusterExtension,
) BuildpackResolver {
	return &buildpackResolver{
		clusterStore:      clusterStore,
		buildpacks:        buildpacks,
		clusterBuildpacks: clusterBuildpacks,
		extensions:        extensions,
		clusterExtensions: clusterExtensions,
	}
}

func (r *buildpackResolver) ClusterStoreObservedGeneration() int64 {
	if r.clusterStore != nil {
		return r.clusterStore.Status.ObservedGeneration
	}
	return 0
}

func (r *buildpackResolver) resolveBuildpack(ref v1alpha2.BuilderBuildpackRef) (K8sRemoteBuildpack, error) {
	var matchingBuildpacks []K8sRemoteBuildpack
	var err error
	switch {
	case ref.Kind == v1alpha2.BuildpackKind && ref.Id != "":
		bp := findBuildpack(ref.ObjectReference, r.buildpacks)
		if bp == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("buildpack not found: %v", ref.Name)
		}

		matchingBuildpacks, err = r.resolveFromBuildpack(ref.Id, []*v1alpha2.Buildpack{bp})
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
	case ref.Kind == v1alpha2.ClusterBuildpackKind && ref.Id != "":
		cbp := findClusterBuildpack(ref.ObjectReference, r.clusterBuildpacks)
		if cbp == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("cluster buildpack not found: %v", ref.Name)
		}

		matchingBuildpacks, err = r.resolveFromClusterBuildpack(ref.Id, []*v1alpha2.ClusterBuildpack{cbp})
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
	case ref.Kind != "":
		bp, err := r.resolveBuildpackFromObjectReference(ref.ObjectReference)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		matchingBuildpacks = []K8sRemoteBuildpack{bp}
	case ref.Id != "":
		bp, err := r.resolveFromBuildpack(ref.Id, r.buildpacks)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		matchingBuildpacks = append(matchingBuildpacks, bp...)

		cbp, err := r.resolveFromClusterBuildpack(ref.Id, r.clusterBuildpacks)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		matchingBuildpacks = append(matchingBuildpacks, cbp...)

		cs, err := r.resolveFromClusterStore(ref.Id, r.clusterStore)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		matchingBuildpacks = append(matchingBuildpacks, cs...)
	case ref.Image != "":
		// TODO(chenbh):
		return K8sRemoteBuildpack{}, fmt.Errorf("using images in builders not currently supported")
	default:
		return K8sRemoteBuildpack{}, fmt.Errorf("invalid buildpack reference")
	}

	if len(matchingBuildpacks) == 0 {
		return K8sRemoteBuildpack{}, errors.Errorf("could not find buildpack with id '%s'", ref.Id)
	}

	if ref.Version == "" {
		bp, err := highestVersion(matchingBuildpacks)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		return bp, nil
	}

	for _, result := range matchingBuildpacks {
		if result.Buildpack.Version == ref.Version {
			return result, nil
		}
	}

	return K8sRemoteBuildpack{}, errors.Errorf("could not find buildpack with id '%s' and version '%s'", ref.Id, ref.Version)
}

func (r *buildpackResolver) resolveExtension(ref v1alpha2.BuilderBuildpackRef) (K8sRemoteBuildpack, error) {
	var (
		matching []K8sRemoteBuildpack
		err      error
	)
	switch {
	case ref.Kind == v1alpha2.ExtensionKind && ref.Id != "":
		ext := findExtension(ref.ObjectReference, r.extensions)
		if ext == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("extension not found: %v", ref.Name)
		}
		matching, err = resolveFromExtension(ref.Id, []*v1alpha2.Extension{ext})
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
	case ref.Kind == v1alpha2.ClusterExtensionKind && ref.Id != "":
		ext := findClusterExtension(ref.ObjectReference, r.clusterExtensions)
		if ext == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("cluster extension not found: %v", ref.Name)
		}
		matching, err = resolveFromClusterExtension(ref.Id, []*v1alpha2.ClusterExtension{ext})
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
	case ref.Kind != "":
		ext, err := r.resolveExtensionFromObjectReference(ref.ObjectReference)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		matching = []K8sRemoteBuildpack{ext}
	case ref.Id != "":
		ext, err := resolveFromExtension(ref.Id, r.extensions)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		matching = append(matching, ext...)
		clusterExt, err := resolveFromClusterExtension(ref.Id, r.clusterExtensions)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		matching = append(matching, clusterExt...)
	case ref.Image != "":
		// TODO: add test
		return K8sRemoteBuildpack{}, fmt.Errorf("using images in builders not currently supported")
	default:
		return K8sRemoteBuildpack{}, fmt.Errorf("invalid extension reference")
	}

	if len(matching) == 0 {
		return K8sRemoteBuildpack{}, errors.Errorf("could not find extension with id '%s'", ref.Id)
	}
	if ref.Version == "" {
		resolved, err := highestVersion(matching)
		if err != nil {
			return K8sRemoteBuildpack{}, err
		}
		return resolved, nil
	}
	for _, result := range matching {
		if result.Buildpack.Version == ref.Version {
			return result, nil
		}
	}
	return K8sRemoteBuildpack{}, errors.Errorf("could not find extension with id '%s' and version '%s'", ref.Id, ref.Version)
}

func (r *buildpackResolver) resolveFromBuildpack(id string, buildpacks []*v1alpha2.Buildpack) ([]K8sRemoteBuildpack, error) {
	var matchingBuildpacks []K8sRemoteBuildpack
	for _, bp := range buildpacks {
		for _, status := range bp.Status.Buildpacks {
			if status.Id == id {
				matchingBuildpacks = append(matchingBuildpacks, K8sRemoteBuildpack{
					Buildpack: status,
					SecretRef: registry.SecretRef{
						ServiceAccount: bp.Spec.ServiceAccountName,
						Namespace:      bp.Namespace,
					},
					source: v1.ObjectReference{Name: bp.Name, Namespace: bp.Namespace, Kind: bp.Kind},
				})
			}
		}
	}
	return matchingBuildpacks, nil
}

func resolveFromExtension(id string, extensions []*v1alpha2.Extension) ([]K8sRemoteBuildpack, error) {
	var matching []K8sRemoteBuildpack
	for _, ext := range extensions {
		for _, status := range ext.Status.Extensions {
			if status.Id == id {
				matching = append(matching, K8sRemoteBuildpack{
					Buildpack: status,
					SecretRef: registry.SecretRef{
						ServiceAccount: ext.Spec.ServiceAccountName,
						Namespace:      ext.Namespace,
					},
					source: v1.ObjectReference{Name: ext.Name, Namespace: ext.Namespace, Kind: ext.Kind},
				})
			}
		}
	}
	return matching, nil
}

func (r *buildpackResolver) resolveFromClusterBuildpack(id string, clusterBuildpacks []*v1alpha2.ClusterBuildpack) ([]K8sRemoteBuildpack, error) {
	var matchingBuildpacks []K8sRemoteBuildpack
	for _, cbp := range clusterBuildpacks {
		for _, status := range cbp.Status.Buildpacks {
			if status.Id == id {
				secretRef := registry.SecretRef{}

				if cbp.Spec.ServiceAccountRef != nil {
					secretRef = registry.SecretRef{
						ServiceAccount: cbp.Spec.ServiceAccountRef.Name,
						Namespace:      cbp.Spec.ServiceAccountRef.Namespace,
					}
				}
				matchingBuildpacks = append(matchingBuildpacks, K8sRemoteBuildpack{
					Buildpack: status,
					SecretRef: secretRef,
					source:    v1.ObjectReference{Name: cbp.Name, Namespace: cbp.Namespace, Kind: cbp.Kind},
				})
			}
		}
	}
	return matchingBuildpacks, nil
}

func resolveFromClusterExtension(id string, extensions []*v1alpha2.ClusterExtension) ([]K8sRemoteBuildpack, error) {
	var matching []K8sRemoteBuildpack
	for _, ext := range extensions {
		for _, status := range ext.Status.Extensions {
			if status.Id == id {
				secretRef := registry.SecretRef{}
				if ext.Spec.ServiceAccountRef != nil {
					secretRef = registry.SecretRef{
						ServiceAccount: ext.Spec.ServiceAccountRef.Name,
						Namespace:      ext.Spec.ServiceAccountRef.Namespace,
					}
				}
				matching = append(matching, K8sRemoteBuildpack{
					Buildpack: status,
					SecretRef: secretRef,
					source:    v1.ObjectReference{Name: ext.Name, Namespace: ext.Namespace, Kind: ext.Kind},
				})
			}
		}
	}
	return matching, nil
}

func (r *buildpackResolver) resolveFromClusterStore(id string, store *v1alpha2.ClusterStore) ([]K8sRemoteBuildpack, error) {
	if store == nil {
		return nil, nil
	}

	var matchingBuildpacks []K8sRemoteBuildpack
	for _, status := range store.Status.Buildpacks {
		if status.Id == id {
			secretRef := registry.SecretRef{}

			if store.Spec.ServiceAccountRef != nil {
				secretRef = registry.SecretRef{
					ServiceAccount: store.Spec.ServiceAccountRef.Name,
					Namespace:      store.Spec.ServiceAccountRef.Namespace,
				}
			}
			matchingBuildpacks = append(matchingBuildpacks, K8sRemoteBuildpack{
				Buildpack: status,
				SecretRef: secretRef,
				source:    v1.ObjectReference{Name: store.Name, Namespace: store.Namespace, Kind: store.Kind},
			})
		}
	}
	return matchingBuildpacks, nil
}

// resolveBuildpackFromObjectReference will get the object and figure out the root
// buildpack by converting it to a buildpack dependency tree
func (r *buildpackResolver) resolveBuildpackFromObjectReference(ref v1.ObjectReference) (K8sRemoteBuildpack, error) {
	var (
		bps       []corev1alpha1.BuildpackStatus
		secretRef registry.SecretRef
		objRef    v1.ObjectReference
	)
	switch ref.Kind {
	case v1alpha2.BuildpackKind:
		bp := findBuildpack(ref, r.buildpacks)
		if bp == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("no buildpack with name '%v'", ref.Name)
		}

		bps = bp.Status.Buildpacks
		objRef = v1.ObjectReference{Name: bp.Name, Namespace: bp.Namespace, Kind: bp.Kind}
		secretRef = registry.SecretRef{
			ServiceAccount: bp.Spec.ServiceAccountName,
			Namespace:      bp.Namespace,
		}
	case v1alpha2.ClusterBuildpackKind:
		cbp := findClusterBuildpack(ref, r.clusterBuildpacks)
		if cbp == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("no cluster buildpack with name '%v'", ref.Name)
		}

		bps = cbp.Status.Buildpacks
		objRef = v1.ObjectReference{Name: cbp.Name, Namespace: cbp.Namespace, Kind: cbp.Kind}
		if cbp.Spec.ServiceAccountRef != nil {
			secretRef = registry.SecretRef{
				ServiceAccount: cbp.Spec.ServiceAccountRef.Name,
				Namespace:      cbp.Spec.ServiceAccountRef.Namespace,
			}
		}
	default:
		return K8sRemoteBuildpack{}, fmt.Errorf("kind must be either %v or %v", v1alpha2.BuildpackKind, v1alpha2.ClusterBuildpackKind)
	}

	trees := NewTree(bps)
	if len(trees) != 1 {
		return K8sRemoteBuildpack{}, fmt.Errorf("unexpected number of root buildpacks: %v", len(trees))
	}

	return K8sRemoteBuildpack{
		Buildpack: *trees[0].Buildpack,
		SecretRef: secretRef,
		source:    objRef,
	}, nil
}

// resolveExtensionFromObjectReference will get the object
func (r *buildpackResolver) resolveExtensionFromObjectReference(ref v1.ObjectReference) (K8sRemoteBuildpack, error) {
	var (
		extensions []corev1alpha1.BuildpackStatus
		secretRef  registry.SecretRef
		objRef     v1.ObjectReference
	)
	switch ref.Kind {
	case v1alpha2.ExtensionKind:
		ext := findExtension(ref, r.extensions)
		if ext == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("no extension with name '%v'", ref.Name)
		}

		extensions = ext.Status.Extensions
		objRef = v1.ObjectReference{Name: ext.Name, Namespace: ext.Namespace, Kind: ext.Kind}
		secretRef = registry.SecretRef{
			ServiceAccount: ext.Spec.ServiceAccountName,
			Namespace:      ext.Namespace,
		}
	case v1alpha2.ClusterExtensionKind:
		ext := findClusterExtension(ref, r.clusterExtensions)
		if ext == nil {
			return K8sRemoteBuildpack{}, fmt.Errorf("no cluster extension with name '%v'", ref.Name)
		}

		extensions = ext.Status.Extensions
		objRef = v1.ObjectReference{Name: ext.Name, Namespace: ext.Namespace, Kind: ext.Kind}
		if ext.Spec.ServiceAccountRef != nil {
			secretRef = registry.SecretRef{
				ServiceAccount: ext.Spec.ServiceAccountRef.Name,
				Namespace:      ext.Spec.ServiceAccountRef.Namespace,
			}
		}
	default:
		return K8sRemoteBuildpack{}, fmt.Errorf("kind must be either %v or %v", v1alpha2.ExtensionKind, v1alpha2.ClusterExtensionKind)
	}

	trees := NewTree(extensions)
	if len(trees) != 1 {
		return K8sRemoteBuildpack{}, fmt.Errorf("unexpected number of root buildpacks: %v", len(trees))
	}

	return K8sRemoteBuildpack{
		Buildpack: *trees[0].Buildpack,
		SecretRef: secretRef,
		source:    objRef,
	}, nil
}

// TODO: combine findBuildpack and findClusterBuildpack into a single func
// if/when golang generics has support for field values
func findBuildpack(ref v1.ObjectReference, buildpacks []*v1alpha2.Buildpack) *v1alpha2.Buildpack {
	for _, bp := range buildpacks {
		if bp.Name == ref.Name {
			return bp
		}
	}
	return nil
}

func findExtension(ref v1.ObjectReference, extensions []*v1alpha2.Extension) *v1alpha2.Extension {
	for _, ext := range extensions {
		if ext.Name == ref.Name {
			return ext
		}
	}
	return nil
}

func findClusterBuildpack(ref v1.ObjectReference, clusterBuildpacks []*v1alpha2.ClusterBuildpack) *v1alpha2.ClusterBuildpack {
	for _, cbp := range clusterBuildpacks {
		if cbp.Name == ref.Name {
			return cbp
		}
	}
	return nil
}

func findClusterExtension(ref v1.ObjectReference, extensions []*v1alpha2.ClusterExtension) *v1alpha2.ClusterExtension {
	for _, ext := range extensions {
		if ext.Name == ref.Name {
			return ext
		}
	}
	return nil
}

// TODO: error if the highest version has multiple diff ids
func highestVersion(matchingBuildpacks []K8sRemoteBuildpack) (K8sRemoteBuildpack, error) {
	for _, bp := range matchingBuildpacks {
		if _, err := semver.NewVersion(bp.Buildpack.Version); err != nil {
			return K8sRemoteBuildpack{}, errors.Errorf("cannot find buildpack '%s' with latest version due to invalid semver '%s'", bp.Buildpack.Id, bp.Buildpack.Version)
		}
	}

	sort.SliceStable(matchingBuildpacks, func(i, j int) bool {
		return semver.MustParse(matchingBuildpacks[i].Buildpack.Version).
			LessThan(semver.MustParse(matchingBuildpacks[j].Buildpack.Version))
	})
	return matchingBuildpacks[len(matchingBuildpacks)-1], nil
}
