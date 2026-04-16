package creatives

import (
	"fmt"
	"net/url"
	"strconv"
)

// GetRequest retrieves a single creative by ID.
type GetRequest struct {
	CreativeID                      string
	IncludeDeletedCreativeSetAssets bool
}

func (r GetRequest) Method() string { return "GET" }
func (r GetRequest) Path() string   { return fmt.Sprintf("/creatives/%s", r.CreativeID) }
func (r GetRequest) Body() any      { return nil }
func (r GetRequest) Query() url.Values {
	v := url.Values{}
	if r.IncludeDeletedCreativeSetAssets {
		v.Set("includeDeletedCreativeSetAssets", strconv.FormatBool(r.IncludeDeletedCreativeSetAssets))
	}
	return v
}
