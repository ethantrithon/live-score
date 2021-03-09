//This Source Code Form is subject to the terms of the Mozilla Public
//License, v. 2.0. If a copy of the MPL was not distributed with this
//file, You can obtain one at http://mozilla.org/MPL/2.0/.

package main

func remove(byteSlice []byte, value byte) []byte {
	ret := []byte{}

	for _, b := range byteSlice {
		if b != value {
			ret = append(ret, b)
		}
	}

	return ret
}

func contains(byteSlice []byte, b byte) bool {
	for _, x := range byteSlice {
		if x == b {
			return true
		}
	}

	return false
}

func unique(byteSlice []byte) []byte {
	ret := []byte{}

	for _, x := range byteSlice {
		if !contains(ret, x) {
			ret = append(ret, x)
		}
	}

	return ret
}
