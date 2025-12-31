package discovery

import (
	"context"
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"github.com/dizzrt/ellie/internal/endpoint"
	"github.com/dizzrt/ellie/log"
	"github.com/dizzrt/ellie/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

var _ resolver.Resolver = (*discoveryResolver)(nil)

type discoveryResolver struct {
	w  registry.Watcher
	cc resolver.ClientConn

	ctx    context.Context
	cancel context.CancelFunc

	insecure    bool
	debugLog    bool
	selectorKey string
	subsetSize  int
}

func (r *discoveryResolver) ResolveNow(_ resolver.ResolveNowOptions) {}

func (r *discoveryResolver) Close() {
	r.cancel()
	err := r.w.Stop()
	if err != nil {
		log.Errorf("[resolver] failed to watch top: %s", err)
	}
}

func parseAttributes(md map[string]string) (attrs *attributes.Attributes) {
	for k, v := range md {
		attrs = attrs.WithValue(k, v)
	}

	return attrs
}

func (r *discoveryResolver) update(ins []*registry.ServiceInstance) {
	var (
		endpoints = make(map[string]struct{})
		filtered  = make([]*registry.ServiceInstance, 0, len(ins))
	)

	for _, in := range ins {
		ept, err := endpoint.Parse(in.Endpoints, endpoint.Scheme("grpc", !r.insecure))
		if err != nil {
			log.Errorf("[resolver] failed to parse discovery endpoint, err: %v", err)
			continue
		}

		if ept == "" {
			continue
		}

		if _, ok := endpoints[ept]; ok {
			continue
		}

		endpoints[ept] = struct{}{}
		filtered = append(filtered, in)
	}

	// if r.subsetSize != 0 {
	// 	filtered = subset.Subset(r.selectorKey, filtered, r.subsetSize)
	// }

	addrs := make([]resolver.Address, 0, len(filtered))
	for _, in := range filtered {
		ept, _ := endpoint.Parse(in.Endpoints, endpoint.Scheme("grpc", !r.insecure))
		addr := resolver.Address{
			ServerName: in.Name,
			Attributes: parseAttributes(in.Metadata).WithValue("rawServiceInstance", in),
			Addr:       ept,
		}

		addrs = append(addrs, addr)
	}

	if len(addrs) == 0 {
		log.Errorf("[resolver] zero endpoint found, refused to write, instances: %v", ins)
		r.cc.ReportError(errors.New("no avaliable service instance found"))
		return
	}

	err := r.cc.UpdateState(resolver.State{Addresses: addrs})
	if err != nil {
		log.Errorf("[resolver] failed to update state: %s", err)
	}

	if r.debugLog {
		b, _ := sonic.Marshal(filtered)
		log.Infof("[resolver] update instances: %s", b)
	}
}

func (r *discoveryResolver) watch() {
	for {
		select {
		case <-r.ctx.Done():
			return
		default:
		}

		ins, err := r.w.Next()
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			log.Errorf("[resolver] failed to watch discovery endpoint: %v", err)
			time.Sleep(time.Second)
			continue
		}

		r.update(ins)
	}
}
