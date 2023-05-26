// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import "syscall"

const (
	AT_REMOVEDIR        = 0x800
	AT_SYMLINK_NOFOLLOW = 0x200

	unlinkatTrap uintptr = syscall.SYS_UNLINKAT
	openatTrap   uintptr = syscall.SYS_OPENAT
)
