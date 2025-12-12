//nolint:stylecheck // generated code and package conventions
package kubemodel

import (
	"time"

	"github.com/google/uuid"
)

type OwnerKind string

const (
	OwnerKindDeployment  OwnerKind = "deployment"
	OwnerKindStatefulSet OwnerKind = "statefulset"
	OwnerKindDaemonSet   OwnerKind = "daemonset"
	OwnerKindJob         OwnerKind = "job"
	OwnerKindCronJob     OwnerKind = "cronjob"
	OwnerKindReplicaSet  OwnerKind = "replicaset"
)

// Owner represents a Kubernetes resource owner (workload controller)
// @bingen:generate:Owner
type Owner struct {
	UID          uuid.UUID         `json:"uid"`                   // @bingen:field[version=1]
	NamespaceUID uuid.UUID         `json:"namespaceUid"`          // @bingen:field[version=1]
	Name         string            `json:"name"`                  // @bingen:field[version=1]
	Kind         OwnerKind         `json:"kind"`                  // @bingen:field[version=1]
	Labels       map[string]string `json:"labels,omitempty"`      // @bingen:field[version=1]
	Annotations  map[string]string `json:"annotations,omitempty"` // @bingen:field[version=1]
	Start        time.Time         `json:"start,omitempty"`       // @bingen:field[version=1] - Owner creation/start timestamp
	End          time.Time         `json:"end,omitempty"`         // @bingen:field[version=1] - Owner deletion/end timestamp (nil if still active)
}
