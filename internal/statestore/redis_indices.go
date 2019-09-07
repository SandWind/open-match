// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package statestore

import (
	"math"
	"strings"

	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/sirupsen/logrus"
	"open-match.dev/open-match/internal/config"
	"open-match.dev/open-match/pkg/pb"
)

func extractIndexedFields(cfg config.View, t *pb.Ticket) map[string]float64 {
	result := make(map[string]float64)

	var indices []string
	if cfg.IsSet("ticketIndices") {
		indices = cfg.GetStringSlice("ticketIndices")
	}

	for _, attribute := range indices {
		v, ok := t.GetProperties().GetFields()[attribute]

		if !ok {
			redisLogger.WithFields(logrus.Fields{
				"attribute": attribute}).Trace("Couldn't find index in Ticket Properties")
			continue
		}

		switch v.Kind.(type) {
		case *structpb.Value_NumberValue:
			result[rangeIndexName(attribute)] = v.GetNumberValue()
		default:
			redisLogger.WithFields(logrus.Fields{
				"attribute": attribute,
			}).Warning("Attribute indexed but is not a number.")
		}
	}

	result[allTickets] = 0

	return result
}

type indexFilter struct {
	name     string
	min, max float64
}

func extractIndexFilters(p *pb.Pool) []indexFilter {
	filters := make([]indexFilter, 0)

	for _, f := range p.FloatRangeFilters {
		filters = append(filters, indexFilter{
			name: rangeIndexName(f.Attribute),
			min:  f.Min,
			max:  f.Max,
		})
	}

	if len(filters) == 0 {
		filters = []indexFilter{{
			name: allTickets,
			min:  math.Inf(-1),
			max:  math.Inf(1),
		}}
	}

	return filters
}

// The following are constants and functions for determining the names of
// indices.  Different index types have different prefixes to avoid any
// name collision.
const allTickets = "allTickets"

func rangeIndexName(attribute string) string {
	// ri stands for range index
	return "ri$" + indexEscape(attribute)
}

func indexCacheName(id string) string {
	// ic stands for index cache
	return "ic$" + indexEscape(id)
}

func indexEscape(s string) string {
	return strings.ReplaceAll(s, "$", "$$")
}