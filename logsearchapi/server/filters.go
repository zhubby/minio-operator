// Copyright (C) 2020, MinIO, Inc.
//
// This code is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License, version 3,
// as published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License, version 3,
// along with this program.  If not, see <http://www.gnu.org/licenses/>

package server

func wildcardRuneMatch(pat, text []rune) bool {
	for len(pat) > 0 {
		switch pat[0] {
		default:
			if len(text) == 0 || text[0] != pat[0] {
				return false
			}
		case '.':
			if len(text) == 0 {
				return false
			}
		case '*':
			return wildcardRuneMatch(pat[1:], text) ||
				(len(text) > 0 && wildcardRuneMatch(pat, text[1:]))
		}
		text = text[1:]
		pat = pat[1:]
	}
	return len(text) == 0 && len(pat) == 0
}

func wildcardMatch(pat, text string) bool {
	return wildcardRuneMatch([]rune(pat), []rune(text))
}

// evalFilters applies the ingest filters on the event and returns if the
// event should be stored.
func (f *IngestFilters) evalFilters(ev *Event) bool {
	// If event matches any exclude filter, return false.
	for _, pat := range f.APINameExclude {
		if wildcardMatch(pat, ev.API.Name) {
			return false
		}
	}

	// If no include filter is given, store the event.
	if len(f.APINameInclude) == 0 {
		return true
	}

	// Otherwise store it only if it matches an include filter.
	for _, pat := range f.APINameInclude {
		if wildcardMatch(pat, ev.API.Name) {
			return true
		}
	}

	return false
}
