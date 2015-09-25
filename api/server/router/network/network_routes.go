package network

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/server/httputils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/daemon"
	"github.com/docker/docker/pkg/parsers/filters"
	"github.com/docker/libnetwork"
)

func (n *networkRouter) getNetworksList(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	filter := r.Form.Get("filters")
	netFilters, err := filters.FromParam(filter)
	if err != nil {
		return err
	}

	list := []*types.NetworkResource{}
	var nameFilter, idFilter bool
	var names, ids []string
	if names, nameFilter = netFilters["name"]; nameFilter {
		for _, name := range names {
			if nw, errRsp := n.daemon.GetNetwork(name, daemon.NetworkByName); errRsp == nil {
				list = append(list, buildNetworkResource(nw))
			}
		}
	}

	if ids, idFilter = netFilters["id"]; idFilter {
		for _, id := range ids {
			for _, nw := range n.daemon.GetNetworksByID(id) {
				list = append(list, buildNetworkResource(nw))
			}
		}
	}

	if !nameFilter && !idFilter {
		nwList := n.daemon.GetNetworksByID("")
		for _, nw := range nwList {
			list = append(list, buildNetworkResource(nw))
		}
	}
	return httputils.WriteJSON(w, http.StatusOK, list)
}

func (n *networkRouter) getNetwork(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	nw, err := n.daemon.FindNetwork(vars["id"])
	if err != nil {
		return err
	}
	return httputils.WriteJSON(w, http.StatusOK, buildNetworkResource(nw))
}

func (n *networkRouter) postNetworkCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	var create types.NetworkCreate
	var warning string

	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		return err
	}

	nw, err := n.daemon.GetNetwork(create.Name, daemon.NetworkByName)
	if _, ok := err.(libnetwork.ErrNoSuchNetwork); err != nil && !ok {
		return err
	}
	if nw != nil {
		if create.CheckDuplicate {
			return libnetwork.NetworkNameError(create.Name)
		}
		warning = fmt.Sprintf("Network with name %s (id : %s) already exists", nw.Name(), nw.ID())
	}

	nw, err = n.daemon.CreateNetwork(create.Name, create.Driver, create.Options)
	if err != nil {
		return err
	}

	return httputils.WriteJSON(w, http.StatusCreated, &types.NetworkCreateResponse{
		ID:      nw.ID(),
		Warning: warning,
	})
}

func (n *networkRouter) postNetworkConnect(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	var connect types.NetworkConnect
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(&connect); err != nil {
		return err
	}

	nw, err := n.daemon.FindNetwork(vars["id"])
	if err != nil {
		return err
	}

	container, err := n.daemon.Get(connect.Container)
	if err != nil {
		return fmt.Errorf("invalid container %s : %v", container, err)
	}
	return container.ConnectToNetwork(nw.Name())
}

func (n *networkRouter) postNetworkDisconnect(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	var disconnect types.NetworkDisconnect
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(&disconnect); err != nil {
		return err
	}

	nw, err := n.daemon.FindNetwork(vars["id"])
	if err != nil {
		return err
	}

	container, err := n.daemon.Get(disconnect.Container)
	if err != nil {
		return fmt.Errorf("invalid container %s : %v", container, err)
	}
	return container.DisconnectFromNetwork(nw)
}

func (n *networkRouter) deleteNetwork(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	nw, err := n.daemon.FindNetwork(vars["id"])
	if err != nil {
		return err
	}

	return nw.Delete()
}

func buildNetworkResource(nw libnetwork.Network) *types.NetworkResource {
	r := &types.NetworkResource{}
	if nw == nil {
		return r
	}

	r.Name = nw.Name()
	r.ID = nw.ID()
	r.Driver = nw.Type()
	r.Containers = make(map[string]types.EndpointResource)
	epl := nw.Endpoints()
	for _, e := range epl {
		sb := e.Info().Sandbox()
		if sb == nil {
			continue
		}

		r.Containers[sb.ContainerID()] = buildEndpointResource(e)
	}
	return r
}

func buildEndpointResource(e libnetwork.Endpoint) types.EndpointResource {
	er := types.EndpointResource{}
	if e == nil {
		return er
	}

	er.EndpointID = e.ID()
	if iface := e.Info().Iface(); iface != nil {
		if mac := iface.MacAddress(); mac != nil {
			er.MacAddress = mac.String()
		}
		if ip := iface.Address(); len(ip.IP) > 0 {
			er.IPv4Address = (&ip).String()
		}

		if ipv6 := iface.AddressIPv6(); len(ipv6.IP) > 0 {
			er.IPv6Address = (&ipv6).String()
		}
	}
	return er
}
