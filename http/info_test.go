// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package http

import (
	"net/http"
	"testing"
)

func TestSetHeader(t *testing.T) {
	tests := []struct {
		name           string
		info           TraceInfo
		expectedInfo   TraceInfo
		expectedParent string
		expectedState  string
	}{
		{
			name: "not sampled, but with trace",
			info: TraceInfo{
				TraceId:  ref(1),
				ParentId: ref(2),
				Sampled:  false,
			},
			expectedInfo: TraceInfo{
				TraceId:  ref(1),
				ParentId: ref(2),
				Sampled:  false,
			},
			expectedParent: "00-0000000000000001-00000002-0",
			expectedState:  "",
		},
		{
			name: "sampled",
			info: TraceInfo{
				TraceId:  ref(1),
				ParentId: ref(16),
				Sampled:  true,
			},
			expectedInfo: TraceInfo{
				TraceId:  ref(1),
				ParentId: ref(16),
				Sampled:  true,
			},
			expectedParent: "00-0000000000000001-00000010-1",
			expectedState:  "",
		},
		{
			name: "sampled without trace",
			info: TraceInfo{
				TraceId:  nil,
				ParentId: ref(16),
				Sampled:  true,
			},
			expectedInfo: TraceInfo{
				TraceId:  nil,
				ParentId: nil,
				Sampled:  true,
			},
			expectedParent: "",
			expectedState:  "sampled=true",
		},
		{
			name: "sampled  with baggage",
			info: TraceInfo{
				TraceId:  ref(1),
				ParentId: ref(16),
				Sampled:  true,
				Baggage: map[string]string{
					"k": "v1",
				},
			},
			expectedInfo: TraceInfo{
				TraceId:  ref(1),
				ParentId: ref(16),
				Sampled:  true,
				Baggage: map[string]string{
					"k": "v1",
				},
			},
			expectedParent: "00-0000000000000001-00000010-1",
			expectedState:  "",
		},
		{
			name: "no trace, no sampled",
			info: TraceInfo{
				TraceId:  nil,
				ParentId: ref(16),
				Sampled:  false,
			},
			expectedInfo: TraceInfo{
				TraceId:  nil,
				ParentId: nil,
				Sampled:  false,
			},
			expectedParent: "",
			expectedState:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			header := http.Header{}
			tc.info.SetHeader(header)
			if header.Get(traceParentHeader) != tc.expectedParent {
				t.Fatalf("%s!=%s", tc.expectedParent, header.Get(traceParentHeader))
			}
			if header.Get(traceStateHeader) != tc.expectedState {
				t.Fatalf("%s!=%s", tc.expectedState, header.Get(traceParentHeader))
			}
			rv := TraceInfoFromHeader(header, "k")
			checkEq(t, tc.expectedInfo.TraceId, rv.TraceId)
			checkEq(t, tc.expectedInfo.ParentId, rv.ParentId)
			if tc.expectedInfo.Sampled != rv.Sampled {
				t.Fatalf("%v!=%v", tc.expectedInfo.Sampled, rv.Sampled)
			}
			for k, v := range rv.Baggage {
				if tc.expectedInfo.Baggage[k] != v {
					t.Fatalf("%v!=%v", tc.expectedInfo.Baggage[k], v)
				}

			}
			for k, v := range tc.expectedInfo.Baggage {
				if rv.Baggage[k] != v {
					t.Fatalf("%v!=%v", rv.Baggage[k], v)
				}
			}
		})

	}

}

func checkEq(t *testing.T, v1 *int64, v2 *int64) {
	if v1 == nil && v2 == nil {
		return
	}
	if v1 == nil || v2 == nil {
		t.Fatalf("One value is nil, other isn't")
	}
	if *v1 != *v2 {
		t.Fatalf("%d!=%d", v1, v2)
	}
}
