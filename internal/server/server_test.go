// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"testing"

	"cloud.google.com/go/bigtable"
	pb "github.com/datacommonsorg/mixer/internal/proto"
)

func TestNoBigTable(t *testing.T) {
	ctx := context.Background()
	s := NewServer(nil, []*bigtable.Table{}, nil)
	_, err := s.GetLandingPageData(ctx, &pb.GetLandingPageDataRequest{
		Place: "geoId/06",
	})
	if err.Error() != "rpc error: code = NotFound desc = Bigtable instance is not specified" {
		t.Errorf("Error invalid: %s", err)
	}
}
