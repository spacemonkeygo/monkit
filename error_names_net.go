// Copyright (C) 2023 Space Monkey, Inc.
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

//go:build !tinygo
// +build !tinygo

package monkit

import (
	"net"
	"os"
)

// getNetErrorName translates net package error.
func getNetErrorName(err error) string {
	switch err.(type) {
	case *os.SyscallError:
	case net.UnknownNetworkError:
		return "Unknown Network Error"
	case *net.AddrError:
		return "Addr Error"
	case net.InvalidAddrError:
		return "Invalid Addr Error"
	case *net.OpError:
		return "Net Op Error"
	case *net.ParseError:
		return "Net Parse Error"
	case *net.DNSError:
		return "DNS Error"
	case *net.DNSConfigError:
		return "DNS Config Error"
	case net.Error:
		return "Network Error"
	}
	return ""
}
