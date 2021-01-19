// Copyright 2018 Google LLC All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

// Resolvable defines the interface for a manifest that can be resolvable to an ImageIndex or an Image.
// Resolves based on mediatype
type Resolvable interface {
	// Image the image this resolves to, or error if it is not an image.
	Image() (Image, error)

	// ImageIndex the image index this resolves to, or error if it is not an image index.
	ImageIndex() (ImageIndex, error)

	// Descriptor the descriptor underlying whatevert his points to
	Descriptor() (Descriptor, error)
}
