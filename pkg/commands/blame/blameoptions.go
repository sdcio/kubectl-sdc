package blame

import (
	"context"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
)

type BlameOptions struct {
	bfc        BlameFilterClient
	namespace  string
	device     string
	pathFilter PathFilters
	filter     BlameFilters
}

type BlameFilterClient interface {
	GetBlameTree(ctx context.Context, namespace string, device string) (*sdcpb.BlameTreeElement, error)
}

func NewBlameOptions(changeBlameFilterClient BlameFilterClient, namespace string, device string, pathFilter PathFilters, filter BlameFilters) *BlameOptions {
	return &BlameOptions{
		bfc:        changeBlameFilterClient,
		namespace:  namespace,
		device:     device,
		pathFilter: pathFilter,
		filter:     filter,
	}
}

// Getters for BlameOptions fields
func (bo *BlameOptions) GetBlameFilterClient() BlameFilterClient {
	return bo.bfc
}

func (bo *BlameOptions) GetNamespace() string {
	return bo.namespace
}

func (bo *BlameOptions) GetDevice() string {
	return bo.device
}

func (bo *BlameOptions) GetPathFilter() PathFilters {
	return bo.pathFilter
}

func (bo *BlameOptions) GetFilter() BlameFilters {
	return bo.filter
}
