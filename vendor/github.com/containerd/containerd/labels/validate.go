/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package labels

import (
	"fmt"

	"github.com/containerd/errdefs"
)

const (
	maxSize = 4096
)

// Validate a label's key and value are under 4096 bytes
func Validate(k, v string) error {
	if (len(k) + len(v)) > maxSize {
		if len(k) > 10 {
			k = k[:10]
		}
		return fmt.Errorf("label key and value greater than maximum size (%d bytes), key: %s: %w", maxSize, k, errdefs.ErrInvalidArgument)
	}
	return nil
}
