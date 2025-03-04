package v4

/*

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/apache/trafficcontrol/lib/go-rfc"
	"github.com/apache/trafficcontrol/lib/go-tc"
	"github.com/apache/trafficcontrol/lib/go-util"
	client "github.com/apache/trafficcontrol/traffic_ops/v4-client"
)

func TestDeliveryServiceServers(t *testing.T) {
	WithObjs(t, []TCObj{CDNs, Types, Tenants, Parameters, Profiles, Statuses, Divisions, Regions, PhysLocations, CacheGroups, Servers, Topologies, ServiceCategories, DeliveryServices}, func() {
		DeleteTestDeliveryServiceServers(t)
		AssignServersToTopologyBasedDeliveryService(t)
		AssignOriginsToTopologyBasedDeliveryServices(t)
		TryToRemoveLastServerInDeliveryService(t)
		AssignServersToNonTopologyBasedDeliveryServiceThatUsesMidTier(t)
		GetTestDSSIMS(t)
		GetTestDSServerByCDN(t)
	})
}

func TestDeliveryServiceServersWithRequiredCapabilities(t *testing.T) {
	WithObjs(t, []TCObj{CDNs, Types, Tenants, Parameters, Profiles, Statuses, Divisions, Regions, PhysLocations, CacheGroups, Servers, ServerCapabilities, Topologies, ServiceCategories, DeliveryServices, DeliveryServicesRequiredCapabilities, ServerServerCapabilities}, func() {
		CreateTestDeliveryServiceServersWithRequiredCapabilities(t)
		CreateTestMSODSServerWithReqCap(t)
	})
}

const dssaTestingXMLID = "test-ds-server-assignments"

func TryToRemoveLastServerInDeliveryService(t *testing.T) {
	opts := client.NewRequestOptions()
	opts.QueryParameters.Set("xmlId", dssaTestingXMLID)
	dses, _, err := TOSession.GetDeliveryServices(opts)
	if err != nil {
		t.Fatalf("Unexpected error trying to get Delivery service with XMLID '%s': %v - alerts: %+v", dssaTestingXMLID, err, dses.Alerts)
	}
	if len(dses.Response) != 1 {
		t.Fatalf("Expected exactly one Delivery service with XMLID '%s', got: %d", dssaTestingXMLID, len(dses.Response))
	}
	ds := dses.Response[0]
	if ds.ID == nil {
		t.Fatalf("Delivery Service '%s' has no ID", dssaTestingXMLID)
	}

	statuses, _, err := TOSession.GetStatuses(client.RequestOptions{})
	if err != nil {
		t.Fatalf("Could not fetch Statuses: %v - alerts: %+v", err, statuses.Alerts)
	}
	if len(statuses.Response) < 1 {
		t.Fatal("Need at least one Status")
	}

	var badStatusID int
	found := false
	for _, status := range statuses.Response {
		if status.Name != tc.CacheStatusOnline.String() && status.Name != tc.CacheStatusReported.String() {
			badStatusID = status.ID
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Need at least one status with a name other than '%s' or '%s'", tc.CacheStatusOnline, tc.CacheStatusReported)
	}

	server := getServer(dssaTestingXMLID, t)
	orgServer := getServer("test-mso-org-01", t)

	resp, _, err := TOSession.CreateDeliveryServiceServers(*ds.ID, []int{*server.ID, *orgServer.ID}, true, client.RequestOptions{})
	if err != nil {
		t.Fatalf("Failed to assign servers to Delivery Service: %v - alerts: %+v", err, resp.Alerts)
	}

	_, _, err = TOSession.CreateDeliveryServiceServers(*ds.ID, []int{*orgServer.ID}, true, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to remove the only EDGE server assigned to a Delivery Service")
	}

	_, _, err = TOSession.CreateDeliveryServiceServers(*ds.ID, []int{*server.ID}, true, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to remove the only ORG server assigned to an MSO-enabled Delivery Service")
	}

	_, _, err = TOSession.DeleteDeliveryServiceServer(*ds.ID, *server.ID, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to remove the only EDGE server assigned to a Delivery Service")
	}

	_, _, err = TOSession.DeleteDeliveryServiceServer(*ds.ID, *orgServer.ID, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to remove the only ORG server assigned to an MSO-enabled Delivery Service")
	}

	alerts, _, err := TOSession.DeleteServer(*server.ID, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to delete the only EDGE server assigned to a Delivery Service")
	} else {
		t.Logf("Got expected error trying to delete the only EDGE server assigned to a Delivery Service: %v - alerts: %+v", err, alerts.Alerts)
	}

	alerts, _, err = TOSession.DeleteServer(*orgServer.ID, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to delete the only ORG server assigned to an MSO-enabled Delivery Service")
	} else {
		t.Logf("Got expected error trying to delete the only ORG server assigned to an MSO-enabled Delivery Service: %v - alerts: %+v", err, alerts.Alerts)
	}

	alerts, _, err = TOSession.AssignDeliveryServiceIDsToServerID(*server.ID, []int{}, true, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to remove a Delivery Service from the only EDGE server to which it is assigned")
	} else {
		t.Logf("Got expected error trying to remove a Delivery Service from the only EDGE server to which it is assigned: %v - alerts: %+v", err, alerts.Alerts)
	}

	alerts, _, err = TOSession.AssignDeliveryServiceIDsToServerID(*orgServer.ID, []int{}, true, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to remove a Delivery Service from the only EDGE server to which it is assigned")
	} else {
		t.Logf("Got expected error trying to remove a Delivery Service from the only EDGE server to which it is assigned: %v - alerts: %+v", err, alerts.Alerts)
	}

	server.StatusID = &badStatusID
	orgServer.StatusID = &badStatusID
	putRequest := tc.ServerPutStatus{
		Status:        util.JSONNameOrIDStr{ID: &badStatusID},
		OfflineReason: util.StrPtr("test"),
	}
	alerts, _, err = TOSession.UpdateServerStatus(*server.ID, putRequest, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to put server into a bad state when it's the only EDGE assigned to a Delivery Service")
	} else {
		t.Logf("Got expected error trying to put server into a bad state when it's the only EDGE assigned to a Delivery Service: %v - alerts: %+v", err, alerts.Alerts)
	}
	alerts, _, err = TOSession.UpdateServerStatus(*orgServer.ID, putRequest, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to put server into a bad state when it's the only ORG assigned to an MSO Delivery Service")
	} else {
		t.Logf("Got expected error trying to put server into a bad state when it's the only ORG assigned to an MSO Delivery Service: %v - alerts: %+v", err, alerts.Alerts)
	}

	alerts, _, err = TOSession.UpdateServer(*server.ID, server, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to put server into a bad state when it's the only EDGE assigned to a Delivery Service")
	} else {
		t.Logf("Got expected error trying to put server into a bad state when it's the only EDGE assigned to a Delivery Service: %v - alerts: %+v", err, alerts.Alerts)
	}
	alerts, _, err = TOSession.UpdateServer(*orgServer.ID, orgServer, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to put server into a bad state when it's the only ORG assigned to an MSO Delivery Service")
	} else {
		t.Logf("Got expected error trying to put server into a bad state when it's the only ORG assigned to an MSO Delivery Service: %v - alerts: %+v", err, alerts.Alerts)
	}

	server.HostName = util.StrPtr(dssaTestingXMLID + "-quest")
	server.ID = nil
	interfaces := server.Interfaces
	for interfaceIndex, i := range interfaces {
		ipAddresses := i.IPAddresses
		for index, ip := range ipAddresses {
			if ip.ServiceAddress {
				str := "100.100.100."
				ip.Address = str + strconv.Itoa(index)
			}
			ipAddresses[index] = ip
		}
		interfaces[interfaceIndex].IPAddresses = ipAddresses
	}
	server.Interfaces = interfaces
	alerts, _, err = TOSession.CreateServer(server, client.RequestOptions{})
	if err != nil {
		t.Fatalf("Failed to create server: %v - alerts: %+v", err, alerts.Alerts)
	}
	server = getServer(*server.HostName, t)

	_, _, err = TOSession.CreateDeliveryServiceServers(*ds.ID, []int{*server.ID}, true, client.RequestOptions{})
	if err == nil {
		t.Error("Didn't get expected error trying to replace the last EDGE server assigned to a Delivery Service with a server in a bad state")
	}

	// Cleanup
	setInactive(t, *ds.ID)
	alerts, _, err = TOSession.DeleteServer(*server.ID, client.RequestOptions{})
	if err != nil {
		t.Errorf("Failed to delete server: %v - alerts: %+v", err, alerts.Alerts)
	}
}

func getServer(hostname string, t *testing.T) tc.ServerV4 {
	opts := client.NewRequestOptions()
	opts.QueryParameters = url.Values{}
	opts.QueryParameters.Set("hostName", hostname)
	servers, _, err := TOSession.GetServers(opts)
	if err != nil {
		t.Fatalf("Unexpected error fetching server '%s': %v - alerts: %+v", hostname, err, servers.Alerts)
	}
	if len(servers.Response) != 1 {
		t.Fatalf("Expected exactly one server with FQDN '%s', got: %d", hostname, len(servers.Response))
	}
	server := servers.Response[0]
	if server.ID == nil {
		t.Fatal("Server had null/undefined ID after creation")
	}
	return server
}

func GetTestDSSIMS(t *testing.T) {
	const noLimit = 999999

	limit := noLimit
	opts := client.NewRequestOptions()
	opts.QueryParameters.Set("limit", strconv.Itoa(limit))
	_, reqInf, err := TOSession.GetDeliveryServiceServers(opts)
	if err != nil {
		t.Errorf("deliveryserviceservers expected: no error, actual: %v", err)
	} else if reqInf.StatusCode != http.StatusOK {
		t.Errorf("expected deliveryserviceservers response code %v, actual %v", http.StatusOK, reqInf.StatusCode)
	}

	opts.Header.Set(rfc.IfModifiedSince, time.Now().UTC().Add(2*time.Second).Format(time.RFC1123))
	_, reqInf, err = TOSession.GetDeliveryServiceServers(opts)
	if err != nil {
		t.Errorf("deliveryserviceservers IMS request expected: no error, actual: %v", err)
	} else if reqInf.StatusCode != http.StatusNotModified {
		t.Errorf("expected deliveryserviceservers IMS response code %v, actual %v", http.StatusNotModified, reqInf.StatusCode)
	}
}

func AssignServersToTopologyBasedDeliveryService(t *testing.T) {
	opts := client.NewRequestOptions()
	opts.QueryParameters.Set("xmlId", "ds-top")
	ds, _, err := TOSession.GetDeliveryServices(opts)
	if err != nil {
		t.Fatalf("cannot get Delivery Service 'ds-top': %v - alerts: %+v", err, ds.Alerts)
	}
	if len(ds.Response) != 1 {
		t.Fatalf("expected one delivery service: 'ds-top', actual: %v", len(ds.Response))
	}
	d := ds.Response[0]
	if d.Topology == nil || d.ID == nil || d.CDNID == nil || d.CDNName == nil {
		t.Fatal("Traffic Ops returned a representation of a Delivery Service that had null or undefined Topology and/or CDN ID and/or CDN Name and/or ID")
	}
	serversResp, _, err := TOSession.GetServers(client.RequestOptions{})
	servers := []tc.ServerV4{}
	for _, s := range serversResp.Response {
		if s.CDNID != nil && *s.CDNID == *d.CDNID && s.Type == tc.CacheTypeEdge.String() {
			servers = append(servers, s)
		}
	}
	if len(servers) < 1 {
		t.Fatalf("expected: at least one EDGE in cdn %s, actual: %v", *d.CDNName, servers)
	}
	if servers[0].ID == nil {
		t.Fatal("Traffic ops returned a representation of a Server that had a null or undefined ID")
	}
	serverNames := []string{}
	for _, s := range servers {
		if s.CDNID != nil && s.HostName != nil && *s.CDNID == *d.CDNID && s.Type == tc.CacheTypeEdge.String() {
			serverNames = append(serverNames, *s.HostName)
		} else {
			t.Fatalf("expected only EDGE servers in cdn '%s', actual: %v", *d.CDNName, servers)
		}
	}
	_, reqInf, err := TOSession.AssignServersToDeliveryService(serverNames, "ds-top", client.RequestOptions{})
	if err == nil {
		t.Fatal("assigning servers to topology-based delivery service - expected: error, actual: nil error")
	}
	if reqInf.StatusCode < http.StatusBadRequest || reqInf.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("assigning servers to topology-based delivery service - expected: 400-level status code, actual: %d", reqInf.StatusCode)
	}

	_, reqInf, err = TOSession.CreateDeliveryServiceServers(*d.ID, []int{*servers[0].ID}, false, client.RequestOptions{})
	if err == nil {
		t.Fatal("creating deliveryserviceserver assignment for topology-based delivery service - expected: error, actual: nil error")
	}
	if reqInf.StatusCode < http.StatusBadRequest || reqInf.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("creating deliveryserviceserver assignment for topology-based delivery service - expected: 400-level status code, actual: %d", reqInf.StatusCode)
	}
}

func AssignOriginsToTopologyBasedDeliveryServices(t *testing.T) {
	// attempt to assign ORG server to a topology-based DS while the ORG server's cachegroup doesn't belong to the topology
	opts := client.NewRequestOptions()
	opts.QueryParameters.Add("hostName", "denver-mso-org-01")
	resp, _, err := TOSession.GetServers(opts)
	if err != nil {
		t.Fatalf("unable to get server 'denver-mso-org-01': %v - alerts: %+v", err, resp.Alerts)
	}
	if len(resp.Response) != 1 {
		t.Fatalf("Expected exactly one server to exist with Host Name 'denver-mso-org-01': %d", len(resp.Response))
	}
	orgServer := resp.Response[0]
	_, reqInf, err := TOSession.AssignServersToDeliveryService([]string{*orgServer.HostName}, "ds-top-req-cap", client.RequestOptions{})
	if err == nil {
		t.Fatal("assigning ORG server to topology-based delivery service while the ORG server's cachegroup does not belong to the topology - expected: error, actual: nil error")
	}
	if reqInf.StatusCode < http.StatusBadRequest || reqInf.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("assigning ORG server to topology-based delivery service while the ORG server's cachegroup does not belong to the topology - expected: 400-level status code, actual: %d", reqInf.StatusCode)
	}
	opts.QueryParameters.Del("hostName")
	opts.QueryParameters.Set("xmlId", "ds-top-req-cap")
	ds, _, err := TOSession.GetDeliveryServices(opts)
	if err != nil {
		t.Fatalf("cannot get Delivery Service 'ds-top-req-cap': %v - alerts: %+v", err, ds.Alerts)
	}
	if len(ds.Response) != 1 {
		t.Fatalf("expected one delivery service: 'ds-top-req-cap', actual: %v", len(ds.Response))
	}
	d := ds.Response[0]
	if d.Topology == nil || d.ID == nil {
		t.Fatal("Traffic Ops returned a representation of a Delivery Service that had null or undefined Topology and/or ID")
	}
	_, reqInf, err = TOSession.CreateDeliveryServiceServers(*d.ID, []int{*orgServer.ID}, false, client.RequestOptions{})
	if err == nil {
		t.Fatal("creating deliveryserviceserver assignment for ORG server to topology-based delivery service while the ORG server's cachegroup does not belong to the topology - expected: error, actual: nil error")
	}
	if reqInf.StatusCode < http.StatusBadRequest || reqInf.StatusCode >= http.StatusInternalServerError {
		t.Fatalf("creating deliveryserviceserver assignment for ORG server to topology-based delivery service while the ORG server's cachegroup does not belong to the topology - expected: 400-level status code, actual: %d", reqInf.StatusCode)
	}

	// attempt to assign ORG server to a topology-based DS while the ORG server's cachegroup belongs to the topology
	assignResp, reqInf, err := TOSession.AssignServersToDeliveryService([]string{*orgServer.HostName}, "ds-top", client.RequestOptions{})
	if err != nil {
		t.Fatalf("assigning Origin server '%s' to Topology-based Delivery Service 'ds-top' while the ORG server's Cache Group belongs to the Topology - expected: no error, actual: %v - alerts: %+v", *orgServer.HostName, err, assignResp.Alerts)
	}
	if reqInf.StatusCode < http.StatusOK || reqInf.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("assigning ORG server to topology-based delivery service while the ORG server's cachegroup belongs to the topology - expected: 200-level status code, actual: %d", reqInf.StatusCode)
	}
	opts.QueryParameters.Set("xmlId", "ds-top")
	ds, _, err = TOSession.GetDeliveryServices(opts)
	if err != nil {
		t.Fatalf("cannot get Delivery Service 'ds-top': %v - alerts: %+v", err, ds.Alerts)
	}
	if len(ds.Response) != 1 {
		t.Fatalf("expected one delivery service: 'ds-top', actual: %v", len(ds.Response))
	}
	d = ds.Response[0]
	if d.Topology == nil || d.ID == nil {
		t.Fatal("Traffic Ops returned a representation of a Delivery Service that had null or undefined Topology and/or ID")
	}
	alerts, reqInf, err := TOSession.CreateDeliveryServiceServers(*d.ID, []int{*orgServer.ID}, true, client.RequestOptions{})
	if err != nil {
		t.Fatalf("assigning Origin server #%d to Topology-based Delivery Service #%d while the server's Cache Group belongs to the Topology - expected: no error, actual: %v - alerts: %+v", *orgServer.ID, *d.ID, err, alerts.Alerts)
	}
	if reqInf.StatusCode < http.StatusOK || reqInf.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("creating deliveryserviceserver assignment for ORG server to topology-based delivery service while the ORG server's cachegroup belongs to the topology - expected: 200-level status code, actual: %d", reqInf.StatusCode)
	}
}

func AssignServersToNonTopologyBasedDeliveryServiceThatUsesMidTier(t *testing.T) {
	opts := client.NewRequestOptions()
	opts.QueryParameters.Set("xmlId", "ds1")
	resp, _, err := TOSession.GetDeliveryServices(opts)
	if err != nil {
		t.Fatalf("cannot get Delivery Service 'ds1': %v - alerts: %+v", err, resp.Alerts)
	}
	if len(resp.Response) != 1 {
		t.Fatalf("expected one delivery service: 'ds1', actual: %d", len(resp.Response))
	}
	dsWithMid := resp.Response[0]
	if dsWithMid.Topology != nil {
		t.Fatal("expected delivery service: 'ds1' to have a nil Topology, actual: non-nil")
	}
	if dsWithMid.CDNID == nil || dsWithMid.CDNName == nil || dsWithMid.ID == nil {
		t.Fatal("Traffic Ops returned a representation of a Delivery Service that had null or undefined CDN ID and/or CDN Name and/or ID")
	}
	serversResp, _, err := TOSession.GetServers(client.RequestOptions{})
	if err != nil {
		t.Fatalf("unable to fetch all servers: %v - alerts: %+v", err, serversResp.Alerts)
	}
	serversIds := []int{}
	for _, s := range serversResp.Response {
		if s.CDNID != nil && *s.CDNID == *dsWithMid.CDNID && s.Type == tc.CacheTypeEdge.String() {
			serversIds = append(serversIds, *s.ID)
		}
	}
	if len(serversIds) < 1 {
		t.Fatalf("expected: at least one EDGE in cdn %s, actual: 0", *dsWithMid.CDNName)
	}

	assignResp, _, err := TOSession.CreateDeliveryServiceServers(*dsWithMid.ID, serversIds, true, client.RequestOptions{})
	if err != nil {
		t.Fatalf("Unexpected error assigning servers %v to Delivery Service #%d: %v - alerts: %+v", serversIds, *dsWithMid.ID, err, assignResp.Alerts)
	}

	opts.QueryParameters = url.Values{"dsId": []string{strconv.Itoa(*dsWithMid.ID)}}
	dsServersResp, _, err := TOSession.GetServers(opts)
	if err != nil {
		t.Fatalf("unable to fetch Delivery Service #%d servers: %v - alerts: %+v", *dsWithMid.ID, err, dsServersResp.Alerts)
	}
	dsServerIds := []int{}
	for _, dss := range dsServersResp.Response {
		dsServerIds = append(dsServerIds, *dss.ID)
	}
	if len(dsServerIds) <= len(serversIds) {
		t.Fatalf("delivery service servers (%d) expected to exceed directly assigned servers (%d) to account for implicitly assigned mid servers", len(dsServerIds), len(serversIds))
	}

	for _, dss := range dsServersResp.Response {
		if dss.CDNID != nil && *dss.CDNID != *dsWithMid.CDNID {
			t.Fatalf("a server for another cdn was returned for this delivery service")
		}
	}
}

func CreateTestDeliveryServiceServersWithRequiredCapabilities(t *testing.T) {
	sscs := testData.ServerServerCapabilities

	testCases := []struct {
		ds          string
		serverName  string
		ssc         tc.ServerServerCapability
		description string
		err         error
		capability  tc.DeliveryServicesRequiredCapability
	}{
		{
			serverName:  "atlanta-edge-01",
			description: "missing requirements for server -> DS assignment",
			err:         errors.New("cannot be assigned to this delivery service"),
			ssc:         sscs[0],
			capability: tc.DeliveryServicesRequiredCapability{
				DeliveryServiceID:  helperGetDeliveryServiceID(t, util.StrPtr("ds-test-minor-versions")),
				RequiredCapability: sscs[1].ServerCapability,
			},
		},
		{
			serverName:  "atlanta-mid-01",
			description: "successful server -> DS assignment",
			err:         nil,
			ssc:         sscs[1],
			capability: tc.DeliveryServicesRequiredCapability{
				DeliveryServiceID:  helperGetDeliveryServiceID(t, util.StrPtr("ds3")),
				RequiredCapability: sscs[1].ServerCapability,
			},
		},
	}

	for _, ctc := range testCases {
		if ctc.capability.DeliveryServiceID == nil || ctc.capability.RequiredCapability == nil {
			t.Errorf("Bad hard-coded test case '%s' - MUST include non-nil DeliveryServiceID and RequiredCapability", ctc.description)
			continue
		}
		t.Run(ctc.description, func(t *testing.T) {
			opts := client.NewRequestOptions()
			opts.QueryParameters.Set("hostName", ctc.serverName)
			resp, _, err := TOSession.GetServers(opts)
			if err != nil {
				t.Fatalf("cannot get Server '%s' by hostname: %v - alerts: %+v", ctc.serverName, err, resp.Alerts)
			}
			servers := resp.Response
			if len(servers) < 1 {
				t.Fatalf("Expected at least one server to exist with Host Name '%s', found none", ctc.serverName)
			}
			server := servers[0]
			if server.ID == nil {
				t.Fatalf("server %s had nil ID", ctc.serverName)
			}

			alerts, _, err := TOSession.CreateDeliveryServicesRequiredCapability(ctc.capability, client.RequestOptions{})
			if err != nil {
				t.Fatalf("Unexpected error creating a relationship between a Delivery Service and a Capability it requires: %v - alerts: %+v", err, alerts.Alerts)
			}

			ctc.ssc.ServerID = server.ID
			sscResp, _, err := TOSession.CreateServerServerCapability(ctc.ssc, client.RequestOptions{})
			if err != nil {
				t.Fatalf("could not associate Capability '%s' to server #%d: %v - alerts: %+v", *ctc.ssc.ServerCapability, *ctc.ssc.ServerID, err, sscResp.Alerts)
			}

			assignResp, _, got := TOSession.CreateDeliveryServiceServers(*ctc.capability.DeliveryServiceID, []int{*server.ID}, true, client.RequestOptions{})
			if ctc.err == nil && got != nil {
				t.Fatalf("Unexpected error creating server-to-Delivery-Service assignments: %v - alerts: %+v", err, assignResp.Alerts)
			} else if ctc.err != nil {
				found := false
				for _, alert := range assignResp.Alerts.Alerts {
					if alert.Level == tc.ErrorLevel.String() && strings.Contains(alert.Text, ctc.err.Error()) {
						found = true
					}
				}
				if !found {
					t.Fatalf("Expected to find an error-level alert relating to '%v', but it wasn't found. Actual alerts: %v", ctc.err, assignResp.Alerts.Alerts)
				}
			}

			alerts, _, err = TOSession.DeleteDeliveryServicesRequiredCapability(*ctc.capability.DeliveryServiceID, *ctc.capability.RequiredCapability, client.RequestOptions{})
			if err != nil {
				t.Fatalf("Unexpected error deleting a relationship between a Delivery Service and a Capability it requires: %v - alerts: %+v", err, alerts.Alerts)
			}
		})
	}
}

func GetTestDSServerByCDN(t *testing.T) {
	opts := client.NewRequestOptions()
	cdns, _, err := TOSession.GetCDNs(opts)
	if err != nil {
		t.Fatalf("unexpected error getting CDNs: %v", err)
	} else if len(cdns.Response) < 2 {
		t.Fatalf("expected at least 2 CDNs but got %d instead", len(cdns.Response))
	}
	dses, _, err := TOSession.GetDeliveryServices(opts)
	if err != nil {
		t.Fatalf("unexpected error getting delivery services: %v", err)
	}
	cdnNonTopologyDSMap := make(map[int][]int)
	dsMap := make(map[int]tc.DeliveryServiceV4)
	for _, ds := range dses.Response {
		dsMap[*ds.ID] = ds
		if ds.Topology != nil && *ds.Topology != "" {
			continue
		}
		cdnNonTopologyDSMap[*ds.CDNID] = append(cdnNonTopologyDSMap[*ds.CDNID], *ds.ID)
	}
	servers, _, err := TOSession.GetServers(opts)
	if err != nil {
		t.Fatalf("unexpected error getting servers: %v", err)
	}
	serverMap := make(map[int]tc.ServerV4)
	cdnEdgeMap := make(map[int][]int)
	for _, server := range servers.Response {
		serverMap[*server.ID] = server
		if !strings.HasPrefix(server.Type, tc.EdgeTypePrefix) {
			continue
		}
		cdnEdgeMap[*server.CDNID] = append(cdnEdgeMap[*server.CDNID], *server.ID)
	}
	emptyCDN := tc.CDN{}
	cdn1 := tc.CDN{}
	cdn2 := tc.CDN{}
	for i, cdn := range cdns.Response {
		if len(cdnNonTopologyDSMap[cdn.ID]) > 0 && len(cdnEdgeMap[cdn.ID]) > 0 {
			if cdn1 == emptyCDN {
				cdn1 = cdns.Response[i]
			} else {
				cdn2 = cdns.Response[i]
				break
			}
		}
	}
	if cdn1 == emptyCDN || cdn2 == emptyCDN {
		t.Fatalf("expected at least 2 CDNs with at least 1 non-topology-based deliveryservice and 1 edge each")
	}
	_, _, err = TOSession.CreateDeliveryServiceServers(cdnNonTopologyDSMap[cdn1.ID][0], []int{cdnEdgeMap[cdn1.ID][0]}, true, client.RequestOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating delivery service servers: %v", err)
	}
	_, _, err = TOSession.CreateDeliveryServiceServers(cdnNonTopologyDSMap[cdn2.ID][0], []int{cdnEdgeMap[cdn2.ID][0]}, true, client.RequestOptions{})
	if err != nil {
		t.Fatalf("unexpected error creating delivery service servers: %v", err)
	}

	opts.QueryParameters.Set("cdn", cdn1.Name)
	opts.QueryParameters.Set("limit", "999999999")
	dss, _, err := TOSession.GetDeliveryServiceServers(opts)
	if err != nil {
		t.Fatalf("unexpected error getting delivery service servers by cdn: %v", err)
	} else if len(dss.Response) < 1 {
		t.Fatalf("getting delivery service servers - expected: at least 1, actual: %d", len(dss.Response))
	}
	for _, d := range dss.Response {
		if *dsMap[*d.DeliveryService].CDNID != cdn1.ID {
			t.Errorf("getting delivery service servers by cdn (%s) - found entry that did not match the given cdn (server hostname = %s, server CDN name = %s, delivery service xmlId = %s, delivery service CDN name = %s", cdn1.Name, *serverMap[*d.Server].HostName, *serverMap[*d.Server].CDNName, *dsMap[*d.DeliveryService].XMLID, *dsMap[*d.DeliveryService].CDNName)
		}
	}
	opts.QueryParameters.Set("cdn", cdn2.Name)
	dss, _, err = TOSession.GetDeliveryServiceServers(opts)
	if err != nil {
		t.Fatalf("unexpected error getting delivery service servers by cdn: %v", err)
	} else if len(dss.Response) < 1 {
		t.Fatalf("getting delivery service servers - expected: at least 1, actual: %d", len(dss.Response))
	}
	for _, d := range dss.Response {
		if *dsMap[*d.DeliveryService].CDNID != cdn2.ID {
			t.Errorf("getting delivery service servers by cdn (%s) - found entry that did not match the given cdn", cdn2.Name)
			t.Errorf("getting delivery service servers by cdn (%s) - found entry that did not match the given cdn (server hostname = %s, server CDN name = %s, delivery service xmlId = %s, delivery service CDN name = %s", cdn2.Name, *serverMap[*d.Server].HostName, *serverMap[*d.Server].CDNName, *dsMap[*d.DeliveryService].XMLID, *dsMap[*d.DeliveryService].CDNName)
		}
	}
}

func CreateTestMSODSServerWithReqCap(t *testing.T) {
	opts := client.NewRequestOptions()
	opts.QueryParameters.Set("xmlID", "msods1")
	dsReqCap, _, err := TOSession.GetDeliveryServicesRequiredCapabilities(opts)
	if err != nil {
		t.Fatalf("Unexpected error retrieving relationships between Delivery Services and Capabilities they require: %v - alerts: %+v", err, dsReqCap.Alerts)
	}

	if len(dsReqCap.Response) == 0 {
		t.Fatal("no delivery service required capabilites found for ds msods1")
	}
	dsrc := dsReqCap.Response[0]
	if dsrc.DeliveryServiceID == nil {
		t.Fatal("Traffic Ops returned a representation of a Delivery Service/Required Capability relationship with null or undefined Delivery Service ID")
	}

	// Associate origin server to msods1 even though it does not have req cap
	// TODO: DON'T hard-code server hostnames!
	opts.QueryParameters = url.Values{}
	opts.QueryParameters.Set("hostName", "denver-mso-org-01")
	resp, _, err := TOSession.GetServers(opts)
	if err != nil {
		t.Fatalf("getting server denver-mso-org-01: %v - alerts: %+v", err, resp.Alerts)
	}
	servers := resp.Response
	if len(servers) != 1 {
		t.Fatal("expected 1 server with hostname denver-mso-org-01")
	}

	s := servers[0]
	if s.ID == nil {
		t.Fatal("server 'denver-mso-org-01' had nil ID")
	}

	// Make sure server has no caps to ensure test correctness
	sccsOpts := client.NewRequestOptions()
	sccsOpts.QueryParameters.Set("serverId", strconv.Itoa(*s.ID))
	sccs, _, err := TOSession.GetServerServerCapabilities(sccsOpts)
	if err != nil {
		t.Fatalf("Unexpected error getting Capabilities for server #%d ('denver-mso-org-01'): %v - alerts: %+v", *s.ID, err, sccs.Alerts)
	}
	if len(sccs.Response) != 0 {
		t.Fatal("expected 0 server server capabilities for server denver-mso-org-01")
	}

	// Is origin included in eligible servers even though it doesnt have required capability
	eServers, _, err := TOSession.GetDeliveryServicesEligible(*dsrc.DeliveryServiceID, client.RequestOptions{})
	if err != nil {
		t.Fatalf("get delivery service msods1 eligible servers: %v - alerts: %+v", err, eServers.Alerts)
	}
	found := false
	for _, es := range eServers.Response {
		if es.HostName != nil && *es.HostName == "denver-mso-org-01" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected to find origin server denver-mso-org-01 to be in eligible server return even though it is missing a required capability")
	}

	if createResp, _, err := TOSession.CreateDeliveryServiceServers(*dsrc.DeliveryServiceID, []int{*s.ID}, true, client.RequestOptions{}); err != nil {
		t.Fatalf("Unexpected error creating server-to-Delivery-Service assignments: %v - alerts: %+v", err, createResp.Alerts)
	}

	// Create new bogus server capability
	if alerts, _, err := TOSession.CreateServerCapability(tc.ServerCapability{Name: "newfun"}, client.RequestOptions{}); err != nil {
		t.Fatalf("cannot create 'newfun' Server Capability: %v - alerts: %+v", err, alerts.Alerts)
	}

	// Attempt to assign to DS should not fail
	if alerts, _, err := TOSession.CreateDeliveryServicesRequiredCapability(tc.DeliveryServicesRequiredCapability{
		DeliveryServiceID:  dsrc.DeliveryServiceID,
		RequiredCapability: util.StrPtr("newfun"),
	}, client.RequestOptions{}); err != nil {
		t.Fatalf("Unexpected error adding Capability 'newfun' as requirement to Delivery Service 'msods1' (#%d): %v - alerts: %+v", *dsrc.DeliveryServiceID, err, alerts.Alerts)
	}

	// Remove required capablity
	if alerts, _, err := TOSession.DeleteDeliveryServicesRequiredCapability(*dsrc.DeliveryServiceID, "newfun", client.RequestOptions{}); err != nil {
		t.Fatalf("Unexpected error removing Capability 'newfun' as requirement from Delivery Service 'msods1' (#%d): %v - alerts: %+v", *dsrc.DeliveryServiceID, err, alerts.Alerts)
	}

	// Delete server capability
	if alerts, _, err := TOSession.DeleteServerCapability("newfun", client.RequestOptions{}); err != nil {
		t.Fatalf("Unexpected error deleteing the 'newfun' Server Capability: %v - alerts: %+v", err, alerts.Alerts)
	}
}

func DeleteTestDeliveryServiceServers(t *testing.T) {
	ds, server := getServerAndDSofSameCDN(t)
	if server.ID == nil {
		t.Fatalf("Got a server with a nil ID: %+v", server)
	}
	if ds.ID == nil {
		t.Fatalf("Got a delivery service with a nil ID %+v", ds)
	}
	if ds.Active == nil {
		t.Fatalf("Got a Delivery Service with nil 'Active': %+v", ds)
	}

	resp, _, err := TOSession.CreateDeliveryServiceServers(*ds.ID, []int{*server.ID}, true, client.RequestOptions{})
	if err != nil {
		t.Fatalf("Unexpected error creating server-to-Delivery-Service assignments: %v - alerts: %+v", err, resp.Alerts)
	}

	dsServers, _, err := TOSession.GetDeliveryServiceServers(client.RequestOptions{})
	if err != nil {
		t.Fatalf("Unexpected error retrieving server-to-Delivery-Service assignments: %v - alerts: %+v", err, dsServers.Alerts)
	}

	found := false
	for _, dss := range dsServers.Response {
		if dss.DeliveryService != nil && *dss.DeliveryService == *ds.ID && dss.Server != nil && *dss.Server == *server.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("POST delivery service servers returned success, but ds-server not in GET")
	}

	serversByDS, _, err := TOSession.GetServersByDeliveryService(*ds.ID, client.RequestOptions{})
	if err != nil {
		t.Errorf("unexpected error getting servers by delivery service: %v", err)
	}
	if len(serversByDS.Response) != 1 {
		t.Errorf("getting servers by delivery service - expected: 1 server, actual: %d servers", len(serversByDS.Response))
	}
	if *serversByDS.Response[0].ID != *server.ID {
		t.Errorf("getting servers by delivery service - expected: server ID %d, actual: %d", *server.ID, *serversByDS.Response[0].ID)
	}

	if *ds.Active {
		*ds.Active = false
		_, _, err = TOSession.UpdateDeliveryService(*ds.ID, ds, client.RequestOptions{})
		if err != nil {
			t.Errorf("Setting Delivery Service #%d to inactive", *ds.ID)
		}
	}

	if alerts, _, err := TOSession.DeleteDeliveryServiceServer(*ds.ID, *server.ID, client.RequestOptions{}); err != nil {
		t.Errorf("Unexpected error removing server-to-Delivery-Service assignments: %v - alerts: %+v", err, alerts.Alerts)
	}

	dsServers, _, err = TOSession.GetDeliveryServiceServers(client.RequestOptions{})
	if err != nil {
		t.Fatalf("Unexpected error retrieving server-to-Delivery-Service assignments: %v - alerts: %+v", err, dsServers.Alerts)
	}

	found = false
	for _, dss := range dsServers.Response {
		if dss.DeliveryService != nil && *dss.DeliveryService == *ds.ID && dss.Server != nil && *dss.Server == *server.ID {
			found = true
			break
		}
	}
	if found {
		t.Error("DELETE delivery service servers returned success, but still in GET")
	}
}

func getServerAndDSofSameCDN(t *testing.T) (tc.DeliveryServiceV4, tc.ServerV4) {
	dses, _, err := TOSession.GetDeliveryServices(client.RequestOptions{})
	if err != nil {
		t.Fatalf("cannot GET DeliveryServices: %v", err)
	}
	if len(dses.Response) < 1 {
		t.Fatal("GET DeliveryServices returned no dses, must have at least 1 to test ds-servers")
	}

	resp, _, err := TOSession.GetServers(client.RequestOptions{})
	if err != nil {
		t.Fatalf("cannot get Servers: %v - alerts: %+v", err, resp.Alerts)
	}
	servers := resp.Response
	if len(servers) < 1 {
		t.Fatal("GET Servers returned no dses, must have at least 1 to test ds-servers")
	}

	for _, ds := range dses.Response {
		for _, s := range servers {
			if ds.CDNName != nil && s.CDNName != nil && *ds.CDNName == *s.CDNName {
				return ds, s
			}
		}
	}
	t.Fatal("expected at least one delivery service and server in the same CDN")

	return tc.DeliveryServiceV4{}, tc.ServerV4{}
}
