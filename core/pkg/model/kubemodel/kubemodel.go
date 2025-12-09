//nolint:stylecheck // generated code and package conventions
package kubemodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// @bingen:generate[stringtable]:KubeModelSet
type KubeModelSet struct {
	Metadata               *Metadata                            `json:"meta"`                   // @bingen:field[version=1]
	Window                 Window                               `json:"window"`                 // @bingen:field[version=1]
	Cluster                *Cluster                             `json:"cluster"`                // @bingen:field[version=1]
	Namespaces             map[uuid.UUID]*Namespace             `json:"namespaces"`             // @bingen:field[version=1]
	ResourceQuotas         map[uuid.UUID]*ResourceQuota         `json:"resourceQuotas"`         // @bingen:field[version=1]
	Containers             map[uuid.UUID]*Container             `json:"containers,omitempty"`   // @bingen:field[version=1]
	Owners                 map[uuid.UUID]*Owner                 `json:"owners,omitempty"`       // @bingen:field[version=1]
	Devices                map[uuid.UUID]*Device                `json:"devices,omitempty"`      // @bingen:field[version=1]
	DeviceUsages           map[uuid.UUID]*DeviceUsage           `json:"deviceUsages,omitempty"` // @bingen:field[version=1]
	Nodes                  map[uuid.UUID]*Node                  `json:"nodes,omitempty"`        // @bingen:field[version=1]
	Pods                   map[uuid.UUID]*Pod                   `json:"pods,omitempty"`         // @bingen:field[version=1]
	PersistentVolumeClaims map[uuid.UUID]*PersistentVolumeClaim `json:"pvcs,omitempty"`         // @bingen:field[version=1]
	Services               map[uuid.UUID]*Service               `json:"services,omitempty"`     // @bingen:field[version=1]
	Volumes                map[uuid.UUID]*PersistentVolume      `json:"volumes,omitempty"`      // @bingen:field[version=1]
	idx                    *kubeModelSetIndexes                 // @bingen:field[ignore]
}

func (kms *KubeModelSet) MarshalBinary() (data []byte, err error) {
	// TODO implement me
	panic("implement me")
}

// UnmarshalJSON implements custom JSON unmarshaling to ensure maps are initialized
func (kms *KubeModelSet) UnmarshalJSON(data []byte) error {
	// Define a type alias to avoid recursion
	type Alias KubeModelSet
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(kms),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Initialize nil maps to empty maps
	if kms.Containers == nil {
		kms.Containers = make(map[uuid.UUID]*Container)
	}
	if kms.Owners == nil {
		kms.Owners = make(map[uuid.UUID]*Owner)
	}
	if kms.Devices == nil {
		kms.Devices = make(map[uuid.UUID]*Device)
	}
	if kms.DeviceUsages == nil {
		kms.DeviceUsages = make(map[uuid.UUID]*DeviceUsage)
	}
	if kms.Namespaces == nil {
		kms.Namespaces = make(map[uuid.UUID]*Namespace)
	}
	if kms.Nodes == nil {
		kms.Nodes = make(map[uuid.UUID]*Node)
	}
	if kms.Pods == nil {
		kms.Pods = make(map[uuid.UUID]*Pod)
	}
	if kms.PersistentVolumeClaims == nil {
		kms.PersistentVolumeClaims = make(map[uuid.UUID]*PersistentVolumeClaim)
	}
	if kms.ResourceQuotas == nil {
		kms.ResourceQuotas = make(map[uuid.UUID]*ResourceQuota)
	}
	if kms.Services == nil {
		kms.Services = make(map[uuid.UUID]*Service)
	}
	if kms.Volumes == nil {
		kms.Volumes = make(map[uuid.UUID]*PersistentVolume)
	}
	if kms.idx == nil {
		kms.idx = &kubeModelSetIndexes{
			namespaceNameToID: make(map[string]uuid.UUID),
		}
	}

	return nil
}

func NewKubeModelSet(start time.Time, end time.Time) *KubeModelSet {
	now := time.Now().UTC()
	kms := &KubeModelSet{
		Metadata: &Metadata{
			Start: now,
			End:   now, // Will be updated when processing completes
		},
		Window: Window{
			Start: start,
			End:   end,
		},
		Containers:             map[uuid.UUID]*Container{},
		Owners:                 map[uuid.UUID]*Owner{},
		Devices:                map[uuid.UUID]*Device{},
		DeviceUsages:           map[uuid.UUID]*DeviceUsage{},
		Namespaces:             map[uuid.UUID]*Namespace{},
		Nodes:                  map[uuid.UUID]*Node{},
		Pods:                   map[uuid.UUID]*Pod{},
		PersistentVolumeClaims: map[uuid.UUID]*PersistentVolumeClaim{},
		ResourceQuotas:         map[uuid.UUID]*ResourceQuota{},
		Services:               map[uuid.UUID]*Service{},
		Volumes:                map[uuid.UUID]*PersistentVolume{},
		idx: &kubeModelSetIndexes{
			namespaceNameToID: map[string]uuid.UUID{},
		},
	}
	// Set the window duration
	if end.After(start) {
		kms.Window.DurationSeconds = uint64(end.Sub(start).Seconds())
	}
	return kms
}

func (kms *KubeModelSet) RegisterNamespace(id string, name string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid namespace UID '%s': %w", id, err)
	}

	if _, ok := kms.Namespaces[uid]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		kms.Namespaces[uid] = &Namespace{
			UID:  uid,
			Name: name,
		}

		// Index namespace name-to-ID for fast lookup
		if name != "" {
			kms.idx.namespaceNameToID[name] = uid
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

// GetNamespaceByName retrieves a namespace by its name using the index
func (kms *KubeModelSet) GetNamespaceByName(name string) (*Namespace, bool) {
	if kms.idx == nil {
		return nil, false
	}

	uid, ok := kms.idx.namespaceNameToID[name]
	if !ok {
		return nil, false
	}

	ns, ok := kms.Namespaces[uid]
	return ns, ok
}

// IsEmpty returns true if the KubeModelSet is nil, has no cluster, or contains no resources
func (kms *KubeModelSet) IsEmpty() bool { //nolint:gocyclo // complexity acceptable for resource checking
	if kms == nil || kms.Cluster == nil {
		return true
	}

	// Check if all resource maps are empty
	return len(kms.Containers) == 0 &&
		len(kms.Owners) == 0 &&
		len(kms.Devices) == 0 &&
		len(kms.DeviceUsages) == 0 &&
		len(kms.Namespaces) == 0 &&
		len(kms.Nodes) == 0 &&
		len(kms.Pods) == 0 &&
		len(kms.PersistentVolumeClaims) == 0 &&
		len(kms.ResourceQuotas) == 0 &&
		len(kms.Services) == 0 &&
		len(kms.Volumes) == 0
}

func (kms *KubeModelSet) RegisterResourceQuota(uid, name, namespace string) error {
	parsedUID, err := uuid.Parse(uid)
	if err != nil {
		return fmt.Errorf("invalid resource quota UID '%s': %w", uid, err)
	}

	if _, ok := kms.ResourceQuotas[parsedUID]; !ok {
		nsUID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.ResourceQuotas[parsedUID] = &ResourceQuota{
			UID:          parsedUID,
			Name:         name,
			NamespaceUID: nsUID,
			Spec:         &ResourceQuotaSpec{Hard: &ResourceQuotaSpecHard{}},
			Status:       &ResourceQuotaStatus{Used: &ResourceQuotaStatusUsed{}},
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterPod(id, name, namespace string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid pod UID '%s': %w", id, err)
	}

	if _, ok := kms.Pods[uid]; !ok {
		nsUID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.Pods[uid] = &Pod{
			UID:          uid,
			Name:         name,
			NamespaceUID: nsUID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterNode(id, name string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid node UID '%s': %w", id, err)
	}

	if _, ok := kms.Nodes[uid]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		kms.Nodes[uid] = &Node{
			UID:             uid,
			Name:            name,
			AttachedVolumes: make(map[uuid.UUID]*NodeVolumeUsage),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterOwner(id, name, namespace, kind string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid owner UID '%s': %w", id, err)
	}

	if _, ok := kms.Owners[uid]; !ok {
		nsUID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.Owners[uid] = &Owner{
			UID:          uid,
			Name:         name,
			NamespaceUID: nsUID,
			Kind:         OwnerKind(kind),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterService(id, name, namespace string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid service UID '%s': %w", id, err)
	}

	if _, ok := kms.Services[uid]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		nsUID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.Services[uid] = &Service{
			UID:          uid,
			ClusterUID:   kms.Cluster.UID,
			NamespaceUID: nsUID,
			Name:         name,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterPVC(id, name, namespace string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid PVC UID '%s': %w", id, err)
	}

	if _, ok := kms.PersistentVolumeClaims[uid]; !ok {
		nsUID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.PersistentVolumeClaims[uid] = &PersistentVolumeClaim{
			UID:          uid,
			Name:         name,
			NamespaceUID: nsUID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterVolume(id, name string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid volume UID '%s': %w", id, err)
	}

	if _, ok := kms.Volumes[uid]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		kms.Volumes[uid] = &PersistentVolume{
			UID:        uid,
			ClusterUID: kms.Cluster.UID,
			Name:       name,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterContainer(id, name, podID string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid container UID '%s': %w", id, err)
	}

	if _, ok := kms.Containers[uid]; !ok {
		podUID, err := uuid.Parse(podID)
		if err != nil {
			return fmt.Errorf("invalid pod UID '%s': %w", podID, err)
		}

		kms.Containers[uid] = &Container{
			PodUID:                    podUID,
			Name:                      name,
			VolumeStorageKiBSeconds:   make(map[uuid.UUID]uint64),
			VolumeStorageByteUsageMax: make(map[uuid.UUID]uint64),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterGPUDevice(id, nodeID string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid device UID '%s': %w", id, err)
	}

	if _, ok := kms.Devices[uid]; !ok {
		nodeUID, err := uuid.Parse(nodeID)
		if err != nil {
			return fmt.Errorf("invalid node UID '%s': %w", nodeID, err)
		}

		kms.Devices[uid] = &Device{
			UID:     uid,
			NodeUID: nodeUID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterGPUUsage(id, containerID, gpuDeviceID string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid device usage UID '%s': %w", id, err)
	}

	if _, ok := kms.DeviceUsages[uid]; !ok {
		containerUID, err := uuid.Parse(containerID)
		if err != nil {
			return fmt.Errorf("invalid container UID '%s': %w", containerID, err)
		}

		deviceUID, err := uuid.Parse(gpuDeviceID)
		if err != nil {
			return fmt.Errorf("invalid device UID '%s': %w", gpuDeviceID, err)
		}

		kms.DeviceUsages[uid] = &DeviceUsage{
			ContainerUID: containerUID,
			DeviceUID:    deviceUID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

type kubeModelSetIndexes struct {
	namespaceNameToID map[string]uuid.UUID
}
