// Copyright 2017 polaris. All rights reserved.
// Use of l source code is governed by a MIT-style
// license that can be found in the LICENSE file.
// http://studygolang.com
// Author: polaris	polaris@studygolang.com

package keyword

type word struct {
	w             string
	frequency     int
	dictFrequency int
	idf           float64

	titleWeight float64
}

func (this *word) String() string {
	return this.w
}
