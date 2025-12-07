//nolint:stylecheck // generated code and package conventions
package kubemodel

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// @bingen:generate[stringtable]:KubeModelSet
type KubeModelSet struct {
	Metadata               *Metadata                         `json:"meta"`                   // @bingen:field[version=1]
	Window                 Window                            `json:"window"`                 // @bingen:field[version=1]
	Cluster                *Cluster                          `json:"cluster"`                // @bingen:field[version=1]
	Namespaces             map[string]*Namespace             `json:"namespaces"`             // @bingen:field[version=1]
	ResourceQuotas         map[string]*ResourceQuota         `json:"resourceQuotas"`         // @bingen:field[version=1]
	Containers             map[string]*Container             `json:"containers,omitempty"`   // @bingen:field[version=1]
	Owners                 map[string]*Owner                 `json:"owners,omitempty"`       // @bingen:field[version=1]
	Devices                map[string]*Device                `json:"devices,omitempty"`      // @bingen:field[version=1]
	DeviceUsages           map[string]*DeviceUsage           `json:"deviceUsages,omitempty"` // @bingen:field[version=1]
	Nodes                  map[string]*Node                  `json:"nodes,omitempty"`        // @bingen:field[version=1]
	Pods                   map[string]*Pod                   `json:"pods,omitempty"`         // @bingen:field[version=1]
	PersistentVolumeClaims map[string]*PersistentVolumeClaim `json:"pvcs,omitempty"`         // @bingen:field[version=1]
	Services               map[string]*Service               `json:"services,omitempty"`     // @bingen:field[version=1]
	Volumes                map[string]*PersistentVolume      `json:"volumes,omitempty"`      // @bingen:field[version=1]
	idx                    *kubeModelSetIndexes              // @bingen:field[ignore]
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
		kms.Containers = make(map[string]*Container)
	}
	if kms.Owners == nil {
		kms.Owners = make(map[string]*Owner)
	}
	if kms.Devices == nil {
		kms.Devices = make(map[string]*Device)
	}
	if kms.DeviceUsages == nil {
		kms.DeviceUsages = make(map[string]*DeviceUsage)
	}
	if kms.Namespaces == nil {
		kms.Namespaces = make(map[string]*Namespace)
	}
	if kms.Nodes == nil {
		kms.Nodes = make(map[string]*Node)
	}
	if kms.Pods == nil {
		kms.Pods = make(map[string]*Pod)
	}
	if kms.PersistentVolumeClaims == nil {
		kms.PersistentVolumeClaims = make(map[string]*PersistentVolumeClaim)
	}
	if kms.ResourceQuotas == nil {
		kms.ResourceQuotas = make(map[string]*ResourceQuota)
	}
	if kms.Services == nil {
		kms.Services = make(map[string]*Service)
	}
	if kms.Volumes == nil {
		kms.Volumes = make(map[string]*PersistentVolume)
	}
	if kms.idx == nil {
		kms.idx = &kubeModelSetIndexes{
			namespaceNameToID: make(map[string]string),
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
		Containers:             map[string]*Container{},
		Owners:                 map[string]*Owner{},
		Devices:                map[string]*Device{},
		DeviceUsages:           map[string]*DeviceUsage{},
		Namespaces:             map[string]*Namespace{},
		Nodes:                  map[string]*Node{},
		Pods:                   map[string]*Pod{},
		PersistentVolumeClaims: map[string]*PersistentVolumeClaim{},
		ResourceQuotas:         map[string]*ResourceQuota{},
		Services:               map[string]*Service{},
		Volumes:                map[string]*PersistentVolume{},
		idx: &kubeModelSetIndexes{
			namespaceNameToID: map[string]string{},
		},
	}
	// Set the window duration
	if end.After(start) {
		kms.Window.DurationSeconds = uint64(end.Sub(start).Seconds())
	}
	return kms
}

func (kms *KubeModelSet) RegisterNamespace(id string, name string) error {
	if _, ok := kms.Namespaces[id]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		kms.Namespaces[id] = &Namespace{
			UID:  id,
			Name: name,
		}

		// Index namespace name-to-ID for fast lookup
		if name != "" {
			kms.idx.namespaceNameToID[name] = id
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

	id, ok := kms.idx.namespaceNameToID[name]
	if !ok {
		return nil, false
	}

	ns, ok := kms.Namespaces[id]
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
	if _, ok := kms.ResourceQuotas[uid]; !ok {
		if _, ok := kms.idx.namespaceNameToID[namespace]; !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.ResourceQuotas[uid] = &ResourceQuota{
			UID:          uid,
			Name:         name,
			NamespaceUID: kms.idx.namespaceNameToID[namespace],
			Spec:         &ResourceQuotaSpec{Hard: &ResourceQuotaSpecHard{}},
			Status:       &ResourceQuotaStatus{Used: &ResourceQuotaStatusUsed{}},
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterPod(id, name, namespace string) error {
	if _, ok := kms.Pods[id]; !ok {
		nsID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.Pods[id] = &Pod{
			UID:          id,
			Name:         name,
			NamespaceUID: nsID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterNode(id, name string) error {
	if _, ok := kms.Nodes[id]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		kms.Nodes[id] = &Node{
			UID:             id,
			Name:            name,
			AttachedVolumes: make(map[string]*NodeVolumeUsage),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterOwner(id, name, namespace, kind string) error {
	if _, ok := kms.Owners[id]; !ok {
		nsID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.Owners[id] = &Owner{
			UID:          id,
			Name:         name,
			NamespaceUID: nsID,
			Kind:         OwnerKind(kind),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterService(id, name, namespace string) error {
	if _, ok := kms.Services[id]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		nsID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.Services[id] = &Service{
			UID:          id,
			ClusterUID:   kms.Cluster.UID,
			NamespaceUID: nsID,
			Name:         name,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterPVC(id, name, namespace string) error {
	if _, ok := kms.PersistentVolumeClaims[id]; !ok {
		nsID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.PersistentVolumeClaims[id] = &PersistentVolumeClaim{
			UID:          id,
			Name:         name,
			NamespaceUID: nsID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterVolume(id, name string) error {
	if _, ok := kms.Volumes[id]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		kms.Volumes[id] = &PersistentVolume{
			UID:        id,
			ClusterUID: kms.Cluster.UID,
			Name:       name,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterContainer(id, name, podID string) error {
	if _, ok := kms.Containers[id]; !ok {
		kms.Containers[id] = &Container{
			PodUID:                    podID,
			Name:                      name,
			VolumeStorageKiBSeconds:   make(map[string]uint64),
			VolumeStorageByteUsageMax: make(map[string]uint64),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterGPUDevice(id, nodeID string) error {
	if _, ok := kms.Devices[id]; !ok {
		kms.Devices[id] = &Device{
			UID:     id,
			NodeUID: nodeID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterGPUUsage(id, containerID, gpuDeviceID string) error {
	if _, ok := kms.DeviceUsages[id]; !ok {
		kms.DeviceUsages[id] = &DeviceUsage{
			ContainerUID: containerID,
			DeviceUID:    gpuDeviceID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

type kubeModelSetIndexes struct {
	namespaceNameToID map[string]string
}
