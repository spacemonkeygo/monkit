// Copyright (C) 2014 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spacemonkeygo/monkit/v3"
	"github.com/spacemonkeygo/monkit/v3/present"
)

const (
	// see: https://github.com/w3c/trace-context/blob/main/spec/20-http_request_header_format.md
	traceSampled      = byte(1)
	traceParentHeader = "traceparent"
	traceStateHeader  = "tracestate"

	// see: https://www.w3.org/TR/baggage/
	baggageHeader = "baggage"

	// see: https://github.com/w3c/trace-context/blob/main/spec/21-http_response_header_format.md
	traceIDHeader = "trace-id"
	childIDHeader = "child-id"

	// orphanSampling is a special k,v which can be added to the vendor specific tracestate header.
	// it can turn on trace sampling on remote even without propagating the parent trace
	// (traceparent must not contain zero IDs)
	// useful when you use curl (no client side tracing), and would like to get traces from the server.
	orphanSampling = "sampled=true"
)

// TraceInfo is a structure representing an incoming RPC request. Every field
// is optional.
type TraceInfo struct {
	TraceId  *int64
	ParentId *int64
	Sampled  bool
	Baggage  map[string]string
}

// HeaderGetter is an interface that http.Header matches for RequestFromHeader
type HeaderGetter interface {
	Get(string) string
}

// HeaderSetter is an interface that http.Header matches for TraceInfo.SetHeader
type HeaderSetter interface {
	Set(string, string)
}

// TraceInfoFromHeader will create a TraceInfo object given a http.Header or
// anything that matches the HeaderGetter interface.
func TraceInfoFromHeader(header HeaderGetter, allowedBaggage ...string) (rv TraceInfo) {
	traceParent := header.Get(traceParentHeader)
	traceState := header.Get(traceStateHeader)
	baggage := header.Get(baggageHeader)

	if traceParent != "" {
		parts := strings.Split(traceParent, "-")
		if len(parts) != 4 {
			return rv
		}
		version, err := hexToUint64(parts[0])
		if err != nil || version != 0 {
			return rv
		}
		traceID, err := hexToUint64(parts[1])
		if err != nil {
			return rv
		}
		parentID, err := hexToUint64(parts[2])
		if err != nil {
			return rv
		}
		flags, err := hexToUint64(parts[3])
		if err != nil {
			return rv
		}

		bm := map[string]string{}
		if baggage != "" {
			for _, kv := range strings.Split(baggage, ",") {
				if key, value, ok := strings.Cut(kv, "="); ok {
					for _, b := range allowedBaggage {
						if key == b {
							bm[key] = value
							break
						}
					}
				}
			}
		}
		return TraceInfo{
			TraceId:  &traceID,
			ParentId: &parentID,
			Sampled:  (byte(flags) & traceSampled) == traceSampled,
			Baggage:  bm,
		}
	}

	// trace parent is not set, but tracing can be turned on by a traceState
	if strings.Contains(traceState, "sampled=true") {
		return TraceInfo{
			Sampled: true,
		}
	}
	return rv
}

func ref(v int64) *int64 {
	return &v
}

func TraceInfoFromSpan(s *monkit.Span) TraceInfo {
	trace := s.Trace()

	sampled, _ := trace.Get(present.SampledKey).(bool)

	if !sampled {
		return TraceInfo{Sampled: sampled}
	}

	req := TraceInfo{
		TraceId:  ref(trace.Id()),
		ParentId: ref(s.Id()),
		Sampled:  sampled,
	}
	if parentID, hasParent := s.ParentId(); hasParent {
		req.ParentId = ref(parentID)
	}
	return req
}

// SetHeader will take a TraceInfo and fill out an http.Header, or anything that
// matches the HeaderSetter interface.
func (r TraceInfo) SetHeader(header HeaderSetter) {
	sampled := byte(0)
	if r.Sampled {
		sampled = traceSampled
	}
	if r.TraceId != nil && r.ParentId != nil {
		header.Set(traceParentHeader, fmt.Sprintf("00-%016x-%08x-%x", *r.TraceId, *r.ParentId, int(sampled)))
	} else if r.Sampled {
		header.Set(traceStateHeader, orphanSampling)
	}

	var baggage []string
	if r.Baggage != nil {
		for k, v := range r.Baggage {
			baggage = append(baggage, fmt.Sprintf("%s=%s", k, v))
		}
		header.Set(baggageHeader, strings.Join(baggage, ","))
	}
}

// hexToUint64 reads a signed int64 that has been formatted as a hex uint64
func hexToUint64(s string) (int64, error) {
	v, err := strconv.ParseUint(s, 16, 64)
	return int64(v), err
}
