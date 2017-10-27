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

	tc "github.com/apache/incubator-trafficcontrol/lib/go-tc"
)

// Tenants gets an array of Tenants
func (to *Session) Tenants() ([]tc.Tenant, error) {
	var data tc.GetTenantsResponse
	err := get(to, tenantsEp(), &data)
	if err != nil {
		return nil, err
	}

	return data.Response, nil
}

// Tenant gets the Tenant for the ID it's passed
func (to *Session) Tenant(id string) (*tc.Tenant, error) {
	var data tc.GetTenantsResponse
	err := get(to, tenantEp(id), &data)
	if err != nil {
		return nil, err
	}

	return &data.Response[0], nil
}

// CreateTenant creates the Tenant it's passed
func (to *Session) CreateTenant(t *tc.Tenant) (*tc.CreateTenantResponse, error) {
	var data tc.CreateTenantResponse
	jsonReq, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	err = post(to, tenantsEp(), jsonReq, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}