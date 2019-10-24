package deliveryservice

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *   http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/apache/trafficcontrol/lib/go-tc"
	"github.com/apache/trafficcontrol/lib/go-tc/tovalidate"
	"github.com/apache/trafficcontrol/lib/go-util"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/api"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/dbhelpers"
	"github.com/apache/trafficcontrol/traffic_ops/traffic_ops_golang/tenant"
	validation "github.com/go-ozzo/ozzo-validation"
)

const (
	deliveryServiceQueryParam    = "deliveryServiceID"
	requiredCapabilityQueryParam = "requiredCapability"
	xmlIDQueryParam              = "xmlID"
)

// RequiredCapability provides a type to define methods on.
type RequiredCapability struct {
	api.APIInfoImpl `json:"-"`
	tc.DeliveryServiceRequiredCapability
}

// SetLastUpdated implements the api.GenericCreator interfaces and
// sets the timestamp on insert.
func (sc *RequiredCapability) SetLastUpdated(t tc.TimeNoMod) { sc.LastUpdated = &t }

// NewReadObj implements the api.GenericReader interfaces.
func (sc *RequiredCapability) NewReadObj() interface{} {
	return &tc.DeliveryServiceRequiredCapability{}
}

// SelectQuery implements the api.GenericReader interface.
func (sc *RequiredCapability) SelectQuery() string {
	return `SELECT
	sc.required_capability,
	sc.deliveryservice_id,
	ds.xml_id,
	sc.last_updated
	FROM deliveryservice_required_capability sc
	JOIN deliveryservice ds ON ds.id = sc.deliveryservice_id`
}

// ParamColumns implements the api.GenericReader interface.
func (sc *RequiredCapability) ParamColumns() map[string]dbhelpers.WhereColumnInfo {
	return map[string]dbhelpers.WhereColumnInfo{
		deliveryServiceQueryParam: dbhelpers.WhereColumnInfo{
			Column:  "sc.deliveryservice_id",
			Checker: api.IsInt,
		},
		xmlIDQueryParam: dbhelpers.WhereColumnInfo{
			Column:  "ds.xml_id",
			Checker: nil,
		},
		requiredCapabilityQueryParam: dbhelpers.WhereColumnInfo{
			Column:  "sc.required_capability",
			Checker: nil,
		},
	}
}

// DeleteQuery implements the api.GenericDeleter interface.
func (sc *RequiredCapability) DeleteQuery() string {
	return `DELETE FROM deliveryservice_required_capability
	WHERE deliveryservice_id = :deliveryservice_id AND required_capability = :required_capability`
}

// GetKeyFieldsInfo implements the api.Identifier interface.
func (sc RequiredCapability) GetKeyFieldsInfo() []api.KeyFieldInfo {
	return []api.KeyFieldInfo{
		{
			Field: deliveryServiceQueryParam,
			Func:  api.GetIntKey,
		},
		{
			Field: requiredCapabilityQueryParam,
			Func:  api.GetStringKey,
		},
	}
}

// GetKeys implements the api.Identifier interface and is not needed
// because Update is not available.
func (sc RequiredCapability) GetKeys() (map[string]interface{}, bool) {
	return nil, false
}

// SetKeys implements the api.Identifier interface and allows the
// create handler to assign deliveryServiceID and requiredCapability.
func (sc *RequiredCapability) SetKeys(keys map[string]interface{}) {
	// this utilizes the non panicking type assertion, if the thrown
	// away ok variable is false it will be the zero of the type.
	id, _ := keys[deliveryServiceQueryParam].(int)
	sc.DeliveryServiceID = &id

	capability, _ := keys[requiredCapabilityQueryParam].(string)
	sc.RequiredCapability = &capability
}

// GetAuditName implements the api.Identifier interface and
// returns the name of the object.
func (sc *RequiredCapability) GetAuditName() string {
	if sc.RequiredCapability != nil {
		return *sc.RequiredCapability
	}

	return "unknown"
}

// GetType implements the api.Identifier interface and
// returns the name of the struct.
func (sc *RequiredCapability) GetType() string {
	return "deliveryservice.RequiredCapability"
}

// Validate implements the api.Validator interface.
func (sc RequiredCapability) Validate() error {
	errs := validation.Errors{
		deliveryServiceQueryParam:    validation.Validate(sc.DeliveryServiceID, validation.Required),
		requiredCapabilityQueryParam: validation.Validate(sc.RequiredCapability, validation.Required),
	}

	return util.JoinErrs(tovalidate.ToErrors(errs))
}

// Update implements the api.CRUDer interface.
func (sc *RequiredCapability) Update() (error, error, int) {
	return nil, nil, http.StatusNotImplemented
}

// Read implements the api.CRUDer interface.
func (sc *RequiredCapability) Read() ([]interface{}, error, error, int) {
	tenantIDs, err := sc.getTenantIDs()
	if err != nil {
		return nil, nil, err, http.StatusInternalServerError
	}

	capabilities, userErr, sysErr, errCode := sc.getCapabilities(tenantIDs)
	if userErr != nil || sysErr != nil {
		return nil, userErr, sysErr, errCode
	}

	results := []interface{}{}
	for _, capability := range capabilities {
		results = append(results, capability)
	}

	return results, nil, nil, http.StatusOK
}

func (sc *RequiredCapability) getTenantIDs() ([]int, error) {
	tenantIDs, err := tenant.GetUserTenantIDListTx(sc.APIInfo().Tx.Tx, sc.APIInfo().User.TenantID)
	if err != nil {
		return nil, err
	}
	return tenantIDs, nil
}

func (sc *RequiredCapability) getCapabilities(tenantIDs []int) ([]tc.DeliveryServiceRequiredCapability, error, error, int) {
	where, orderBy, pagination, queryValues, errs := dbhelpers.BuildWhereAndOrderByAndPagination(sc.APIInfo().Params, sc.ParamColumns())
	if len(errs) > 0 {
		return nil, util.JoinErrs(errs), nil, http.StatusBadRequest
	}

	where, queryValues = dbhelpers.AddTenancyCheck(where, queryValues, "ds.tenant_id", tenantIDs)
	query := sc.SelectQuery() + where + orderBy + pagination

	rows, err := sc.APIInfo().Tx.NamedQuery(query, queryValues)
	if err != nil {
		return nil, nil, err, http.StatusInternalServerError
	}
	defer rows.Close()

	var results []tc.DeliveryServiceRequiredCapability
	for rows.Next() {
		var result tc.DeliveryServiceRequiredCapability
		if err := rows.StructScan(&result); err != nil {
			return nil, nil, errors.New(sc.GetType() + " get scanning: " + err.Error()), http.StatusInternalServerError
		}
		results = append(results, result)
	}

	return results, nil, nil, 0
}

// Delete implements the api.CRUDer interface.
func (sc *RequiredCapability) Delete() (error, error, int) {
	authorized, err := sc.isTenantAuthorized()
	if err != nil {
		return nil, errors.New("checking tenant: " + err.Error()), http.StatusInternalServerError
	} else if !authorized {
		return errors.New("not authorized on this tenant"), nil, http.StatusForbidden
	}

	return api.GenericDelete(sc)
}

// Create implements the api.CRUDer interface.
func (sc *RequiredCapability) Create() (error, error, int) {
	authorized, err := sc.isTenantAuthorized()
	if err != nil {
		return nil, errors.New("checking tenant: " + err.Error()), http.StatusInternalServerError
	} else if !authorized {
		return errors.New("not authorized on this tenant"), nil, http.StatusForbidden
	}

	rows, err := sc.APIInfo().Tx.NamedQuery(scInsertQuery(), sc)
	if err != nil {
		return api.ParseDBError(err)
	}
	defer rows.Close()

	rowsAffected := 0
	for rows.Next() {
		rowsAffected++
		if err := rows.StructScan(&sc); err != nil {
			return nil, errors.New(sc.GetType() + " create scanning: " + err.Error()), http.StatusInternalServerError
		}
	}
	if rowsAffected == 0 {
		return nil, errors.New(sc.GetType() + " create: no " + sc.GetType() + " was inserted, no rows was returned"), http.StatusInternalServerError
	} else if rowsAffected > 1 {
		return nil, errors.New("too many rows returned from " + sc.GetType() + " insert"), http.StatusInternalServerError
	}

	return nil, nil, http.StatusOK
}

func (sc *RequiredCapability) isTenantAuthorized() (bool, error) {
	if sc.DeliveryServiceID == nil && sc.XMLID == nil {
		return false, errors.New("delivery service has no ID or XMLID")
	}

	var existingID *int
	var err error

	switch {
	case sc.DeliveryServiceID != nil:
		existingID, _, err = getDSTenantIDByID(sc.APIInfo().Tx.Tx, *sc.DeliveryServiceID)
		if err != nil {
			return false, err
		}
	case sc.XMLID != nil:
		existingID, _, err = getDSTenantIDByName(sc.APIInfo().Tx.Tx, tc.DeliveryServiceName(*sc.XMLID))
		if err != nil {
			return false, err
		}
	}

	if existingID != nil {
		authorized, err := tenant.IsResourceAuthorizedToUserTx(*existingID, sc.APIInfo().User, sc.APIInfo().Tx.Tx)
		if err != nil {
			return false, fmt.Errorf("checking authorization for existing DS ID: %s" + err.Error())
		}
		if !authorized {
			return false, errors.New("not authorized on this tenant")
		}
	}

	return true, err
}

func scInsertQuery() string {
	return `INSERT INTO deliveryservice_required_capability (
required_capability,
deliveryservice_id) VALUES (
:required_capability,
:deliveryservice_id) RETURNING deliveryservice_id, required_capability, last_updated`
}