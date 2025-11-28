package types

import (
	"strconv"
	"time"
)

type PermissionType string

const (
	PushRegistryPermission    PermissionType = "PushRegistry"
	PullRegistryPermission    PermissionType = "PullRegistry"
	PushExperimentsPermission PermissionType = "PushExperiments"
)

type Permissions struct {
	PullRegistry    bool
	PushRegistry    bool
	PushExperiments bool
	Expiry          time.Duration
}

func NewPermissions(m map[string]string) (Permissions, error) {
	pullRegistry, err := strconv.ParseBool(m["PullRegistry"])
	if err != nil {
		return Permissions{}, err
	}

	pushRegistry, err := strconv.ParseBool(m["PushRegistry"])
	if err != nil {
		return Permissions{}, err
	}

	pushExperiments, err := strconv.ParseBool(m["PushExperiments"])
	if err != nil {
		return Permissions{}, err
	}

	return Permissions{
		PullRegistry:    pullRegistry,
		PushRegistry:    pushRegistry,
		PushExperiments: pushExperiments,
	}, nil
}

func (p *Permissions) Mapping() map[string]string {
	return map[string]string{
		"PullRegistry":    strconv.FormatBool(p.PullRegistry),
		"PushRegistry":    strconv.FormatBool(p.PushRegistry),
		"PushExperiments": strconv.FormatBool(p.PushExperiments),
	}
}

func (p *Permissions) HasPermission(perm PermissionType) bool {
	switch perm {
	case PullRegistryPermission:
		return p.PullRegistry
	case PushRegistryPermission:
		return p.PushRegistry
	case PushExperimentsPermission:
		return p.PushExperiments
	default:
		return false
	}
}
