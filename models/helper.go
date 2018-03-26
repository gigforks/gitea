// Copyright 2017 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package models

import "reflect"

func keysInt64(m map[int64]struct{}) []int64 {
	var keys = make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func valuesRepository(m map[int64]*Repository) []*Repository {
	var values = make([]*Repository, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func getSlicesIntersection(a interface{}, b interface{}) []interface{} {
	set := make([]interface{}, 0)
	hash := make(map[interface{}]bool)
	av := reflect.ValueOf(a)
	bv := reflect.ValueOf(b)

	for i := 0; i < av.Len(); i++ {
		el := av.Index(i).Interface()
		hash[el] = true
	}

	for i := 0; i < bv.Len(); i++ {
		el := bv.Index(i).Interface()
		if _, found := hash[el]; found {
			set = append(set, el)
		}
	}

	return set
}

// subset returns true if the first array is completely
// contained in the second array. There must be at least
// the same number of duplicate values in second as there
// are in first.
func subset(first, second []string) bool {
	set := make(map[string]struct{})
	for _, value := range second {
		set[value] = struct {}{}
	}

	for _, value := range first {
		if _, found := set[value]; !found {
			return false
		}
	}

	return true
}
