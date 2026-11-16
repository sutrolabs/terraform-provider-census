package provider

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSourceSourceRead_StorageBackend(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		backend interface{}
		want    map[string]interface{}
	}{
		"s3": {
			backend: map[string]interface{}{"type": "s3", "bucket_name": "my-census-bucket", "role_arn": "arn:aws:iam::123456789012:role/census-role"},
			want:    map[string]interface{}{"type": "s3", "bucket_name": "my-census-bucket", "role_arn": "arn:aws:iam::123456789012:role/census-role"},
		},
		"none": {
			backend: nil,
			want:    map[string]interface{}{},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			apiClient := newSourceTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/sources/2280673" {
					http.NotFound(w, r)
					return
				}
				resp := buildSourceResponse(2280673, "advanced")
				resp["data"].(map[string]interface{})["group_id"] = "wandering_aimlessly"
				resp["data"].(map[string]interface{})["storage_backend"] = tc.backend
				writeJSON(t, w, http.StatusOK, resp)
			})

			d := schema.TestResourceDataRaw(t, dataSourceSource().Schema, map[string]interface{}{"id": "2280673"})
			if diags := dataSourceSourceRead(context.Background(), d, apiClient); diags.HasError() {
				t.Fatalf("expected data source read to succeed, got diagnostics: %#v", diags)
			}

			if got := d.Get("storage_backend").(map[string]interface{}); !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("unexpected storage_backend: got %#v, want %#v", got, tc.want)
			}
			if got := d.Get("group_id").(string); got != "wandering_aimlessly" {
				t.Fatalf("expected group_id to be read, got %q", got)
			}
		})
	}
}
