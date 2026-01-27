// Copyright (C) 2026
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

package monkit

import "time"

type SpanPluginFinisher interface {
	Finish(span *Span, err error, panicked bool, done time.Time)
}

type SpanPlugin interface {
	Start(span *Span) SpanPluginFinisher
}

func (r *Registry) AddSpanPlugin(h SpanPlugin) {
	for {
		pluginsPtr := r.plugins.Load()
		if pluginsPtr == nil {
			r.plugins.CompareAndSwap(nil, new([]SpanPlugin))
			continue
		}
		plugins := *pluginsPtr

		nextPlugins := make([]SpanPlugin, len(plugins)+1)
		copy(nextPlugins, plugins)
		nextPlugins[len(plugins)] = h

		if r.plugins.CompareAndSwap(pluginsPtr, &nextPlugins) {
			return
		}
	}
}

func (r *Registry) getSpanPlugins() []SpanPlugin {
	if pluginsPtr := r.plugins.Load(); pluginsPtr != nil {
		return *pluginsPtr
	}
	return nil
}
