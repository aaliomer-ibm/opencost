package kubemodel

import (
	"errors"
	"fmt"
	"time"
)

// TODO: should we add a lock so that we can safely modify KubeModelSet in parallel?

// @bingen:generate[stringtable]:KubeModelSet
type KubeModelSet struct {
	Metadata               *Metadata                         `json:"meta"`                   // @bingen:field[version=1]
	Window                 Window                            `json:"window"`                 // @bingen:field[version=1]
	Cluster                *Cluster                          `json:"cluster"`                // @bingen:field[version=1]
	Namespaces             map[string]*Namespace             `json:"namespaces"`             // @bingen:field[version=1]
	ResourceQuotas         map[string]*ResourceQuota         `json:"resourceQuotas"`         // @bingen:field[version=1]
	Containers             map[string]*Container             `json:"containers,omitempty"`   // @bingen:field[ignore]
	Owners                 map[string]*Owner                 `json:"owners,omitempty"`       // @bingen:field[ignore]
	Devices                map[string]*Device                `json:"devices,omitempty"`      // @bingen:field[ignore]
	DeviceUsages           map[string]*DeviceUsage           `json:"deviceUsages,omitempty"` // @bingen:field[ignore]
	Nodes                  map[string]*Node                  `json:"nodes,omitempty"`        // @bingen:field[ignore]
	Pods                   map[string]*Pod                   `json:"pods,omitempty"`         // @bingen:field[ignore]
	PersistentVolumeClaims map[string]*PersistentVolumeClaim `json:"pvcs,omitempty"`         // @bingen:field[ignore]
	Services               map[string]*Service               `json:"services,omitempty"`     // @bingen:field[ignore]
	Volumes                map[string]*PersistentVolume      `json:"volumes,omitempty"`      // @bingen:field[ignore]
	idx                    *kubeModelSetIndexes              // @bingen:field[ignore]
}

func NewKubeModelSet(start time.Time, end time.Time) *KubeModelSet {
	now := time.Now().UTC()
	kms := &KubeModelSet{
		Metadata: &Metadata{
			CreatedAt:   now,
			CompletedAt: now, // Will be updated when processing completes
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
		idx:                    newKubeModelSetIndexes(),
	}
	// Set the window duration
	if end.After(start) {
		kms.Window.DurationSeconds = Measurement(end.Sub(start).Seconds())
	}
	return kms
}

func (kms *KubeModelSet) RegisterNamespace(uid string, name string) error {
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
func (kms *KubeModelSet) IsEmpty() bool {
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

func (kms *KubeModelSet) RegisterResourceQuota(uid string, name, namespace string) error {
	if _, ok := kms.ResourceQuotas[uid]; !ok {
		nsUID, ok := kms.idx.namespaceNameToID[namespace]
		if !ok {
			return fmt.Errorf("KubeModelSet missing namespace '%s'", namespace)
		}

		kms.ResourceQuotas[uid] = &ResourceQuota{
			UID:          uid,
			Name:         name,
			NamespaceUID: nsUID,
			Spec:         &ResourceQuotaSpec{Hard: &ResourceQuotaSpecHard{}},
			Status:       &ResourceQuotaStatus{Used: &ResourceQuotaStatusUsed{}},
		}

	}

	return nil
}

func (kms *KubeModelSet) RegisterPod(uid string, name, namespace string) error {
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

func (kms *KubeModelSet) RegisterNode(uid string, name string) error {
	if _, ok := kms.Nodes[uid]; !ok {
		if kms.Cluster == nil {
			return errors.New("KubeModelSet missing Cluster")
		}

		kms.Nodes[uid] = &Node{
			UID:             uid,
			Name:            name,
			AttachedVolumes: make(map[string]*NodeVolumeUsage),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterOwner(uid string, name, namespace, kind string) error {
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

func (kms *KubeModelSet) RegisterService(uid string, name, namespace string) error {
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

func (kms *KubeModelSet) RegisterPVC(uid string, name, namespace string) error {
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

func (kms *KubeModelSet) RegisterVolume(uid string, name string) error {
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

func (kms *KubeModelSet) RegisterContainer(uid string, name string, podUID string) error {
	if _, ok := kms.Containers[uid]; !ok {
		kms.Containers[uid] = &Container{
			PodUID:                    podUID,
			Name:                      name,
			VolumeStorageByteSeconds:  make(map[string]Measurement),
			VolumeStorageByteUsageMax: make(map[string]Measurement),
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterDevice(uid string, nodeUID string) error {
	if _, ok := kms.Devices[uid]; !ok {
		kms.Devices[uid] = &Device{
			UID:     uid,
			NodeUID: nodeUID,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

func (kms *KubeModelSet) RegisterUsage(id, containerID, deviceId string) error {

	if _, ok := kms.DeviceUsages[deviceId]; !ok {

		kms.DeviceUsages[deviceId] = &DeviceUsage{
			ContainerUID: containerID,
			DeviceUID:    deviceId,
		}

		kms.Metadata.ObjectCount++
	}

	return nil
}

type kubeModelSetIndexes struct {
	namespaceNameToID map[string]string
	namespaceByName   map[string]*Namespace
}

func newKubeModelSetIndexes() *kubeModelSetIndexes {
	return &kubeModelSetIndexes{
		namespaceNameToID: make(map[string]string),
		namespaceByName:   make(map[string]*Namespace),
	}
}
