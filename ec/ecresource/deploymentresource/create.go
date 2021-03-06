// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package deploymentresource

import (
	"context"
	"fmt"

	"github.com/elastic/cloud-sdk-go/pkg/api"
	"github.com/elastic/cloud-sdk-go/pkg/api/deploymentapi"
	"github.com/elastic/cloud-sdk-go/pkg/multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Create will create a new deployment from the specified settings.
func Create(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*api.API)
	reqID := deploymentapi.RequestID(d.Get("request_id").(string))

	req, err := createResourceToModel(d)
	if err != nil {
		diag.FromErr(err)
	}

	res, err := deploymentapi.Create(deploymentapi.CreateParams{
		API:       client,
		RequestID: reqID,
		Request:   req,
		Overrides: &deploymentapi.PayloadOverrides{
			Name:    d.Get("name").(string),
			Version: d.Get("version").(string),
			Region:  d.Get("region").(string),
		},
	})
	if err != nil {
		merr := multierror.NewPrefixed("failed creating deployment", err)
		return diag.FromErr(merr.Append(newCreationError(reqID)))
	}

	if err := WaitForPlanCompletion(client, *res.ID); err != nil {
		merr := multierror.NewPrefixed("failed tracking create progress", err)
		return diag.FromErr(merr.Append(newCreationError(reqID)))
	}

	d.SetId(*res.ID)

	if diag := Read(ctx, d, meta); diag != nil {
		return diag
	}

	if err := parseCredentials(d, res.Resources); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func newCreationError(reqID string) error {
	return fmt.Errorf(
		`set "request_id" to "%s" to recreate the deployment resources`, reqID,
	)
}
