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

package client

import (
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/apache/trafficcontrol/lib/go-tc"
)

const (
	v14DeliveryServiceRequiredCapabilities = apiBase + "/deliveryservice_required_capabilities"
)

// CreateDeliveryServiceRequiredCapability assigns a Required Capability to a Delivery Service
func (to *Session) CreateDeliveryServiceRequiredCapability(capability tc.DeliveryServiceRequiredCapability) (tc.Alerts, ReqInf, error) {
	var alerts tc.Alerts
	var remoteAddr net.Addr
	reqInf := ReqInf{CacheHitStatus: CacheHitStatusMiss, RemoteAddr: remoteAddr}
	reqBody, err := json.Marshal(capability)
	if err != nil {
		return tc.Alerts{}, reqInf, err
	}
	reqInf, err = post(to, v14DeliveryServiceRequiredCapabilities, reqBody, &alerts)
	return alerts, reqInf, err
}

// DeleteDeliveryServiceRequiredCapability unassigns a Required Capability from a Delivery Service
func (to *Session) DeleteDeliveryServiceRequiredCapability(deliveryserviceID int, capability string) (tc.Alerts, ReqInf, error) {
	var alerts tc.Alerts
	param := url.Values{}
	param.Add("deliveryServiceID", strconv.Itoa(deliveryserviceID))
	param.Add("requiredCapability", capability)
	url := fmt.Sprintf("%s?%s", v14DeliveryServiceRequiredCapabilities, param.Encode())
	reqInf, err := del(to, url, &alerts)
	return alerts, reqInf, err
}

// GetDeliveryServiceRequiredCapabilities retrieves a list of Required Capabilities that are assigned to a Delivery Service
// Callers can filter the results by delivery service id, xml id and/or required capability via the optional parameters
func (to *Session) GetDeliveryServiceRequiredCapabilities(deliveryServiceID *int, xmlID, capability *string) ([]tc.DeliveryServiceRequiredCapability, ReqInf, error) {
	param := url.Values{}
	if deliveryServiceID != nil {
		param.Add("deliveryServiceID", strconv.Itoa(*deliveryServiceID))
	}
	if xmlID != nil {
		param.Add("xmlID", *xmlID)
	}
	if capability != nil {
		param.Add("requiredCapability", *capability)
	}

	url := v14DeliveryServiceRequiredCapabilities
	if len(param) > 0 {
		url = fmt.Sprintf("%s?%s", url, param.Encode())
	}

	resp := struct {
		Response []tc.DeliveryServiceRequiredCapability `json:"response"`
	}{}
	reqInf, err := get(to, url, &resp)
	if err != nil {
		return nil, reqInf, err
	}
	return resp.Response, reqInf, nil
}